<script lang="ts">
  import { untrack } from 'svelte';
  import type { ControlDTO, ProfileDTO } from '../lib/api';
  import ControlsTable from './ControlsTable.svelte';
  import CalibCopy from './CalibCopy.svelte';
  import CopyModal from './CopyModal.svelte';

  let {
    profile,
    profiles,
    controls,
    onEditControl,
  }: {
    profile: ProfileDTO;
    profiles: ProfileDTO[];
    controls: ControlDTO[];
    onEditControl: (name: string) => void;
  } = $props();

  let tab = $state<'controls' | 'calib'>('controls');
  let search = $state('');
  let selectedControls = $state<Record<string, boolean>>({});
  let copyOpen = $state(false);
  let hint = $state('');

  // Calibration targets, keyed by profile id, defaulting to all except source.
  const defaultCalibSelection = () =>
    Object.fromEntries(profiles.filter((p) => p.id !== profile.id).map((p) => [p.id, true]));

  let selectedProfiles = $state<Record<string, boolean>>(untrack(defaultCalibSelection));

  const selectedNames = $derived(
    controls.filter((c) => selectedControls[c.name]).map((c) => c.name),
  );
  const boundCount = $derived(controls.filter((c) => c.kind !== 'unbound').length);
  const copyableCount = $derived(
    controls.filter((c) => c.kind === 'key' || c.kind === 'button').length,
  );

  const profileKey = () => `${profile.id}|${profiles.map((p) => p.id).join(',')}`;
  let lastProfileKey = untrack(profileKey);
  $effect(() => {
    const key = profileKey();
    if (key !== lastProfileKey) {
      lastProfileKey = key;
      selectedProfiles = defaultCalibSelection();
      selectedControls = {};
      search = '';
      hint = '';
    }
  });

  function toggleControl(name: string) {
    selectedControls = { ...selectedControls, [name]: !selectedControls[name] };
  }

  function toggleAllControls(names: string[]) {
    const all = names.length > 0 && names.every((n) => selectedControls[n]);
    const sel = { ...selectedControls };
    for (const n of names) sel[n] = !all;
    selectedControls = sel;
  }

  function toggleProfile(id: string) {
    selectedProfiles = { ...selectedProfiles, [id]: !selectedProfiles[id] };
  }

  function copySelected() {
    if (selectedNames.length === 0) {
      hint = 'Select at least one control to copy.';
      return;
    }
    hint = '';
    copyOpen = true;
  }
</script>

<div class="profile-page">
  <header class="profile-header">
    <div class="profile-title">
      <span class="eyebrow">Profile</span>
      <h2>{profile.displayName}</h2>
      <p class="meta">
        {controls.length} controls
        <span aria-hidden="true">·</span>
        {boundCount} bound
        <span aria-hidden="true">·</span>
        {copyableCount} copyable
      </p>
    </div>
  </header>

  <nav class="tabs" aria-label="Profile views">
    <button class:active={tab === 'controls'} onclick={() => (tab = 'controls')}>Controls</button>
    <button class:active={tab === 'calib'} onclick={() => (tab = 'calib')}>Calibration</button>
  </nav>

  {#if tab === 'controls'}
    <ControlsTable
      {controls}
      selected={selectedControls}
      {search}
      onSearch={(v) => (search = v)}
      onToggle={toggleControl}
      onToggleAll={toggleAllControls}
      onEdit={onEditControl}
      onCopySelected={copySelected}
    />
    {#if hint}<p class="hint">{hint}</p>{/if}
  {:else}
    <CalibCopy
      {profiles}
      selected={selectedProfiles}
      source={profile.id}
      onToggle={toggleProfile}
    />
  {/if}

  {#if copyOpen}
    <CopyModal
      {profiles}
      sourceId={profile.id}
      controls={selectedNames}
      onClose={() => (copyOpen = false)}
    />
  {/if}
</div>

<style>
  .profile-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .profile-header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
  }
  .profile-title {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .eyebrow {
    color: var(--faint);
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.09em;
    text-transform: uppercase;
  }
  h2 {
    font-size: 1.45rem;
    font-weight: 750;
    letter-spacing: -0.025em;
  }
  .meta {
    color: var(--muted);
    font-size: 0.8rem;
  }
  .meta span {
    margin: 0 0.25rem;
    color: var(--faint);
  }
  .tabs {
    display: inline-flex;
    align-self: flex-start;
    gap: 0.2rem;
    padding: 0.2rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--surface);
  }
  .tabs button {
    padding: 0.38rem 0.8rem;
    border: 0;
    border-radius: 999px;
    background: transparent;
    color: var(--muted);
    font-size: 0.8rem;
    font-weight: 600;
  }
  .tabs button:hover {
    color: var(--text);
  }
  .tabs button.active {
    background: var(--accent);
    color: #fff;
    font-weight: 700;
  }
  .hint {
    color: var(--warning);
    font-size: 0.82rem;
  }
  @media (max-width: 640px) {
    .profile-header {
      align-items: flex-start;
    }
    h2 {
      font-size: 1.2rem;
    }
  }
</style>
