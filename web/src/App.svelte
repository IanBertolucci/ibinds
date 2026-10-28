<script lang="ts">
  import Sidebar from './components/Sidebar.svelte';
  import ProfilePage from './components/ProfilePage.svelte';
  import BindingModal from './components/BindingModal.svelte';
  import Settings from './components/Settings.svelte';
  import Icon from './components/Icon.svelte';
  import { api, type ControlDTO, type ProfileDTO } from './lib/api';

  let profiles = $state<ProfileDTO[]>([]);
  let activeId = $state('');
  let activeControls = $state<ControlDTO[]>([]);
  let controlNames = $state<ControlDTO[]>([]);
  let bindingModal = $state<{ preset: string } | null>(null);
  let settingsOpen = $state(false);
  let error = $state('');
  let profilesDir = $state('');

  const activeProfile = $derived(profiles.find((p) => p.id === activeId));

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
    <div class="brand">
      <span class="brand-mark">iB</span>
      <div class="brand-copy">
        <h1>iBinds</h1>
        <span>iRacing controls</span>
      </div>
    </div>
    <div class="topbar-actions">
      <button class="btn btn-primary" onclick={() => openBindingModal()}>New binding</button>
      <button class="icon-btn" aria-label="Settings" onclick={() => (settingsOpen = true)}>
        <Icon name="settings" />
      </button>
    </div>
  </header>

  {#if error}
    <div class="notices">
      <div class="banner banner-error" role="alert">{error}</div>
    </div>
  {/if}

  <div class="layout">
    <aside class="rail">
      <Sidebar {profiles} {activeId} onSelect={selectProfile} />
    </aside>

    <section class="content">
      {#if activeProfile}
        <ProfilePage
          profile={activeProfile}
          {profiles}
          controls={activeControls}
          onEditControl={(name) => openBindingModal(name)}
        />
      {:else}
        <div class="empty-state">
          <h2>No profile selected</h2>
          <p>Choose a profile in the sidebar to inspect its controls.</p>
        </div>
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

  {#if settingsOpen}
    <Settings
      profilesDir={profilesDir}
      onSaved={(d) => {
        profilesDir = d;
        loadProfiles();
      }}
      onClose={() => (settingsOpen = false)}
    />
  {/if}
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }
  .topbar {
    position: sticky;
    top: 0;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    height: 62px;
    padding: 0 20px;
    border-bottom: 1px solid var(--border);
    background: rgba(255, 255, 255, 0.92);
    backdrop-filter: blur(8px);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 0.65rem;
  }
  .brand-mark {
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    border-radius: 8px;
    background: var(--accent);
    color: #fff;
    font-size: 0.76rem;
    font-weight: 800;
    letter-spacing: 0.01em;
  }
  .brand-copy {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
  }
  .brand-copy h1 {
    font-size: 1.05rem;
    font-weight: 750;
    letter-spacing: -0.02em;
  }
  .brand-copy span {
    color: var(--muted);
    font-size: 0.78rem;
  }
  .topbar-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .notices {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
    max-width: 1600px;
    margin: 0.9rem auto 0;
    padding: 0 20px;
  }
  .layout {
    display: grid;
    grid-template-columns: 280px minmax(0, 1fr);
    gap: 20px;
    width: 100%;
    max-width: 1600px;
    margin: 0 auto;
    padding: 20px;
    flex: 1;
  }
  .rail {
    min-width: 0;
    align-self: start;
    position: sticky;
    top: 82px;
    max-height: calc(100vh - 102px);
    overflow: auto;
  }
  .content {
    min-width: 0;
  }
  .empty-state {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 2rem;
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface);
    color: var(--muted);
  }
  .empty-state h2 {
    color: var(--text);
    font-size: 1rem;
  }
  @media (max-width: 1100px) {
    .layout {
      grid-template-columns: 1fr;
    }
    .rail {
      position: static;
      max-height: none;
    }
  }
  @media (max-width: 640px) {
    .topbar {
      padding: 0 12px;
    }
    .brand-copy span {
      display: none;
    }
    .layout {
      padding: 12px;
      gap: 12px;
    }
    .notices {
      padding: 0 12px;
    }
  }
</style>
