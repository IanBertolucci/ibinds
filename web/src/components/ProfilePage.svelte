<script lang="ts">
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

  // Calibration targets, keyed by profile id, defaulting to all except source.
  let selectedProfiles = $state<Record<string, boolean>>(
    Object.fromEntries(profiles.filter((p) => p.id !== profile.id).map((p) => [p.id, true])),
  );

  const selectedNames = $derived(
    controls.filter((c) => selectedControls[c.name]).map((c) => c.name),
  );

  function toggleControl(name: string) {
    selectedControls = { ...selectedControls, [name]: !selectedControls[name] };
  }

  function toggleAllControls() {
    const copyable = controls.filter((c) => c.kind === 'key' || c.kind === 'button');
    const all = copyable.length > 0 && copyable.every((c) => selectedControls[c.name]);
    const sel: Record<string, boolean> = {};
    for (const c of copyable) sel[c.name] = !all;
    selectedControls = sel;
  }

  function toggleProfile(id: string) {
    selectedProfiles = { ...selectedProfiles, [id]: !selectedProfiles[id] };
  }
</script>

<div class="profile-page">
  <h2>{profile.displayName}</h2>

  <nav>
    <button onclick={() => (tab = 'controls')} class:active={tab === 'controls'}>Controls</button>
    <button onclick={() => (tab = 'calib')} class:active={tab === 'calib'}>Calibration</button>
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
      onCopySelected={() => (copyOpen = true)}
    />
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
  nav {
    display: flex;
    gap: 0.5rem;
  }
  button.active {
    font-weight: bold;
  }
</style>
