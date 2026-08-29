package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	sessionstate "github.com/marang/sway-title-animator/internal/session"
	"github.com/marang/sway-title-animator/internal/statefile"
	"github.com/marang/sway-title-animator/internal/swayipc"
)

const (
	sessionSnapshotDebounce   = time.Second
	sessionObservationDelay   = 2 * time.Second
	sessionStartupSettleDelay = 10 * time.Second
	sessionStartupRetryDelay  = 5 * time.Second
)

type swayRequester interface {
	Request(swayipc.MessageType, []byte) (swayipc.Message, error)
}

type sessionRuntime struct {
	client          swayRequester
	root            string
	persisted       sessionstate.LayoutSnapshot
	desired         sessionstate.LayoutSnapshot
	debouncer       *sessionstate.SnapshotDebouncer
	registryPresent bool
	startupComplete bool
	startupDeadline time.Time
	observeDeadline time.Time
	shutdown        bool
}

type sessionErrorReporter struct {
	lastMessage string
	lastAt      time.Time
}

func (reporter *sessionErrorReporter) Report(err error) {
	if err == nil {
		return
	}
	message := err.Error()
	if message == reporter.lastMessage && time.Since(reporter.lastAt) < 5*time.Second {
		return
	}
	reporter.lastMessage = message
	reporter.lastAt = time.Now()
	fmt.Fprintf(os.Stderr, "Unable to update persistent Sway session: %v\n", err)
}

func newSessionRuntime(client swayRequester) (*sessionRuntime, error) {
	root, err := sessionstate.DefaultStateRoot()
	if err != nil {
		return nil, err
	}
	previous := sessionstate.LayoutSnapshot{
		Version:    sessionstate.LayoutSchemaVersion,
		Workspaces: []sessionstate.WorkspaceLayout{},
	}
	if err := sessionstate.LayoutFile(root).LoadInto(&previous); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load persistent Sway layout: %w", err)
	}
	registry := sessionstate.Registry{}
	registryErr := sessionstate.RegistryFile(root).LoadInto(&registry)
	registryPresent := registryErr == nil
	if registryErr != nil && !errors.Is(registryErr, os.ErrNotExist) {
		return nil, fmt.Errorf("load persistent context registry: %w", registryErr)
	}
	debouncer, err := sessionstate.NewSnapshotDebouncer(previous, sessionSnapshotDebounce)
	if err != nil {
		return nil, fmt.Errorf("initialize Sway layout debounce: %w", err)
	}
	return &sessionRuntime{
		client:          client,
		root:            root,
		persisted:       previous,
		desired:         previous,
		debouncer:       debouncer,
		registryPresent: registryPresent,
		startupComplete: len(previous.Workspaces) == 0,
	}, nil
}

// Reconcile applies placement for newly mapped stable application IDs. It
// returns true only when the caller must obtain a fresh tree before capture.
func (runtime *sessionRuntime) Reconcile(root *Node, now time.Time) (bool, error) {
	if runtime == nil || runtime.shutdown {
		return false, nil
	}
	registry, available, err := runtime.loadRegistry()
	if err != nil {
		runtime.observeDeadline = now.Add(sessionStartupRetryDelay)
		return false, err
	}
	if !available {
		runtime.registryPresent = false
		runtime.observeDeadline = time.Time{}
		runtime.startupDeadline = time.Time{}
		return false, err
	}
	runtime.registryPresent = true
	if !runtime.startupComplete && runtime.startupDeadline.IsZero() {
		runtime.startupDeadline = now.Add(sessionStartupSettleDelay)
	}
	// Sway does not emit an IPC event for every geometry change (notably
	// resize). Keep a low-frequency semantic observation active only while a
	// persistent registry exists so those changes still reach the debouncer.
	runtime.observeDeadline = now.Add(sessionObservationDelay)
	actions, err := sessionstate.PlanPlacementActions(root, registry, runtime.desired)
	if err != nil {
		return false, err
	}
	if len(actions) != 0 {
		for _, action := range actions {
			if err := runtime.applyPlacementAction(action); err != nil {
				var unknown *swayipc.CommandOutcomeUnknownError
				var invalid *swayipc.CommandResponseInvalidError
				return errors.As(err, &unknown) || errors.As(err, &invalid), err
			}
		}
		return true, nil
	}

	captured, err := sessionstate.CaptureLayout(root, registry)
	if err != nil {
		return false, err
	}
	if !runtime.startupComplete {
		ready, err := sessionstate.StartupCaptureReady(runtime.desired, captured, registry)
		if err != nil {
			return false, err
		}
		if !ready && now.Before(runtime.startupDeadline) {
			return false, nil
		}
		runtime.startupComplete = true
		runtime.startupDeadline = time.Time{}
	}
	stable, err := sessionstate.PreserveMissingPlacements(runtime.persisted, captured, registry)
	if err != nil {
		return false, err
	}
	runtime.desired = stable
	_, err = runtime.debouncer.Observe(stable, now)
	return false, err
}

func (runtime *sessionRuntime) loadRegistry() (sessionstate.Registry, bool, error) {
	registry := sessionstate.Registry{}
	err := sessionstate.RegistryFile(runtime.root).LoadInto(&registry)
	if errors.Is(err, os.ErrNotExist) {
		return sessionstate.Registry{}, false, nil
	}
	if err != nil {
		return sessionstate.Registry{}, false, fmt.Errorf("load persistent context registry: %w", err)
	}
	return registry, true, nil
}

func (runtime *sessionRuntime) applyPlacementAction(action sessionstate.PlacementAction) error {
	var command string
	switch action.Kind {
	case sessionstate.PlacementMoveWorkspace:
		command = fmt.Sprintf(
			"[con_id=%d] move container to workspace %s",
			action.ContainerID,
			quoteSwayString(action.Workspace),
		)
	case sessionstate.PlacementAddMark:
		mark, err := action.ContextID.Mark()
		if err != nil {
			return fmt.Errorf("derive persistent Sway mark: %w", err)
		}
		command = fmt.Sprintf("[con_id=%d] mark --add %s", action.ContainerID, quoteSwayString(mark))
	default:
		return fmt.Errorf("unsupported placement action %q", action.Kind)
	}
	message, err := runtime.client.Request(swayipc.RunCommand, []byte(command))
	if err != nil {
		return fmt.Errorf("apply %s for context %q: %w", action.Kind, action.ContextID, err)
	}
	if err := swayipc.CheckRunCommandResponse(message); err != nil {
		return fmt.Errorf("apply %s for context %q: %w", action.Kind, action.ContextID, err)
	}
	return nil
}

func (runtime *sessionRuntime) Deadline() (time.Time, bool) {
	if runtime == nil || runtime.shutdown {
		return time.Time{}, false
	}
	deadline, scheduled := runtime.debouncer.Deadline()
	if !runtime.observeDeadline.IsZero() &&
		(!scheduled || runtime.observeDeadline.Before(deadline)) {
		deadline = runtime.observeDeadline
		scheduled = true
	}
	if !runtime.startupComplete && !runtime.startupDeadline.IsZero() &&
		(!scheduled || runtime.startupDeadline.Before(deadline)) {
		return runtime.startupDeadline, true
	}
	return deadline, scheduled
}

func (runtime *sessionRuntime) ObservationDue(now time.Time) bool {
	return runtime != nil && !runtime.shutdown && !runtime.observeDeadline.IsZero() &&
		!now.Before(runtime.observeDeadline)
}

func (runtime *sessionRuntime) ArmObservationRetry(now time.Time) {
	if runtime != nil && !runtime.shutdown && runtime.registryPresent {
		runtime.observeDeadline = now.Add(sessionObservationDelay)
	}
}

func (runtime *sessionRuntime) PostponeObservation(now time.Time) {
	if runtime != nil && !runtime.shutdown && !runtime.observeDeadline.IsZero() {
		runtime.observeDeadline = now.Add(sessionObservationDelay)
	}
}

func (runtime *sessionRuntime) StartupDue(now time.Time) bool {
	return runtime != nil && !runtime.shutdown && !runtime.startupComplete &&
		!runtime.startupDeadline.IsZero() && !now.Before(runtime.startupDeadline)
}

func (runtime *sessionRuntime) PostponeStartup(now time.Time) {
	if runtime != nil && !runtime.shutdown && !runtime.startupComplete {
		runtime.startupDeadline = now.Add(sessionStartupRetryDelay)
	}
}

func (runtime *sessionRuntime) Flush(now time.Time) error {
	if runtime == nil || runtime.shutdown {
		return nil
	}
	candidate, due := runtime.debouncer.Due(now)
	if !due {
		return nil
	}
	err := sessionstate.LayoutFile(runtime.root).Save(candidate)
	if err == nil {
		runtime.persisted = candidate
		return runtime.debouncer.MarkPersisted(candidate)
	}

	var unknown *statefile.CommitOutcomeUnknownError
	if errors.As(err, &unknown) {
		var visible sessionstate.LayoutSnapshot
		if loadErr := sessionstate.LayoutFile(runtime.root).LoadInto(&visible); loadErr != nil {
			runtime.debouncer.Postpone(now)
			return errors.Join(err, fmt.Errorf("reload layout after unknown commit outcome: %w", loadErr))
		}
		candidateHash, candidateErr := sessionstate.SemanticSnapshotHash(candidate)
		visibleHash, visibleErr := sessionstate.SemanticSnapshotHash(visible)
		if candidateErr != nil || visibleErr != nil || candidateHash != visibleHash {
			runtime.debouncer.Postpone(now)
			return errors.Join(err, candidateErr, visibleErr, errors.New("visible layout differs from the candidate after unknown commit outcome"))
		}
		runtime.desired = visible
		runtime.persisted = visible
		if markErr := runtime.debouncer.MarkPersisted(visible); markErr != nil {
			return errors.Join(err, markErr)
		}
		return err
	}
	runtime.debouncer.Postpone(now)
	return err
}

func (runtime *sessionRuntime) Shutdown() {
	if runtime == nil {
		return
	}
	runtime.shutdown = true
	runtime.startupDeadline = time.Time{}
	runtime.observeDeadline = time.Time{}
	runtime.debouncer.Cancel()
}

func reconcilePersistentSession(animator *TitleAnimator, runtime *sessionRuntime, phase int, report func(error)) {
	const maximumObservations = 4
	if runtime != nil {
		runtime.ArmObservationRetry(time.Now())
	}
	for range maximumObservations {
		root, err := animator.RefreshTree(phase)
		if err != nil {
			// An IPC disconnect preserves the last snapshot. The normal event
			// reconnect path will obtain another tree without turning a socket
			// outage into persistent diagnostic noise.
			return
		}
		if runtime == nil {
			return
		}
		refresh, err := runtime.Reconcile(root, time.Now())
		if err != nil && report != nil {
			report(err)
		}
		if !refresh {
			return
		}
	}
	if report != nil {
		report(errors.New("persistent Sway placement did not stabilize after bounded re-observation"))
	}
}
