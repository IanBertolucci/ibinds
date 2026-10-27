<script lang="ts">
  import Sidebar from './components/Sidebar.svelte';
  import ProfilePage from './components/ProfilePage.svelte';
  import BindingModal from './components/BindingModal.svelte';
  import Settings from './components/Settings.svelte';
  import { api, type ControlDTO, type ProfileDTO } from './lib/api';

  let profiles = $state<ProfileDTO[]>([]);
  let activeId = $state('');
  let activeControls = $state<ControlDTO[]>([]);
  let controlNames = $state<ControlDTO[]>([]);
  let bindingModal = $state<{ preset: string } | null>(null);
  let message = $state('');
  let error = $state('');
  let profilesDir = $state('');

  async function loadProfiles() {
    try {
      profiles = await api.fetchProfiles();
      if (profiles.length > 0 && !activeId) {
        activeId = profiles[0].id;
      }
      // Shared control-name list for the wizard picker (all profiles share
      // the same control names; use the first profile that has controls.cfg).
      for (const p of profiles) {
        if (!p.hasControls) continue;
        try {
          controlNames = await api.fetchControls(p.id);
        } catch {
          /* ignore */
        }
        break;
      }
    } catch (e) {
      error = String(e);
    }
  }

  async function loadActiveControls() {
    if (!activeId) {
      activeControls = [];
      return;
    }
    try {
      activeControls = await api.fetchControls(activeId);
    } catch {
      activeControls = [];
    }
  }

  function selectProfile(id: string) {
    activeId = id;
  }

  function openBindingModal(preset = '') {
    bindingModal = { preset };
  }

  $effect(() => {
    loadProfiles();
  });

  $effect(() => {
    loadActiveControls();
  });
</script>

<main>
  <header class="topbar">
    <h1>iBinds</h1>
    <button onclick={() => openBindingModal()}>New binding</button>
  </header>

  {#if error}<p class="error">{error}</p>{/if}
  {#if message}<p class="ok">{message}</p>{/if}

  <div class="layout">
    <aside>
      <Sidebar {profiles} {activeId} onSelect={selectProfile} />
      <Settings profilesDir={profilesDir} onSaved={(d) => (profilesDir = d)} />
    </aside>

    <section class="content">
      {#if activeId}
        {#each profiles as p (p.id)}
          {#if p.id === activeId}
            <ProfilePage
              profile={p}
              {profiles}
              controls={activeControls}
              onEditControl={(name) => openBindingModal(name)}
            />
          {/if}
        {/each}
      {/if}
    </section>
  </div>

  {#if bindingModal}
    <BindingModal
      {profiles}
      controls={controlNames}
      presetControl={bindingModal.preset}
      onClose={() => (bindingModal = null)}
    />
  {/if}
</main>

<style>
  main {
    font-family: system-ui, sans-serif;
    max-width: 80rem;
    margin: 0 auto;
    padding: 1rem;
  }
  .topbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .layout {
    display: grid;
    grid-template-columns: 16rem 1fr;
    gap: 2rem;
    margin-top: 1rem;
  }
  .error {
    color: #b00;
  }
  .ok {
    color: #080;
  }
</style>
