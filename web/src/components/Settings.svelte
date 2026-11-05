<script lang="ts">
  import { untrack } from 'svelte';
  import { api, type DeviceInfo, type Settings as SettingsDTO } from '../lib/api';
  import { deriveProfilesDir } from '../lib/store';
  import Icon from './Icon.svelte';
  import { focusTrap } from '../lib/focusTrap';

  let {
    iRacingDir = '',
    profilesDir,
    defaultDeviceInstanceGUID,
    firstRun = false,
    onSaved,
    onClose,
  }: {
    iRacingDir?: string;
    profilesDir: string;
    defaultDeviceInstanceGUID: string;
    firstRun?: boolean;
    onSaved: (settings: SettingsDTO) => void;
    onClose: () => void;
  } = $props();

  let value = $state(untrack(() => iRacingDir));
  let deviceValue = $state(untrack(() => defaultDeviceInstanceGUID));
  let devices = $state<DeviceInfo[]>([]);
  let message = $state('');
  let error = $state('');

  const derivedProfilesDir = $derived(value.trim() ? deriveProfilesDir(value) : profilesDir);

  $effect(() => {
    api
      .fetchInputDevices()
      .then((d) => {
        devices = d;
      })
      .catch(() => {
        /* offline fallback is optional */
      });
  });

  async function save() {
    message = '';
    error = '';
    try {
      const res = await api.saveSettings(value.trim(), deviceValue);
      message = 'Saved.';
      onSaved(res);
    } catch (e) {
      error = String(e);
    }
  }
</script>

<div
  class="overlay"
  role="presentation"
  onclick={(e) => {
    if (!firstRun && e.target === e.currentTarget) onClose();
  }}
>
  <div
    class="modal settings-modal"
    use:focusTrap={firstRun ? () => {} : onClose}
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    aria-labelledby="settings-title"
  >
    <div class="modal-header">
      <div>
        <h2 id="settings-title">
          {firstRun ? 'Confirm your iRacing folder' : 'Settings'}
        </h2>
        {#if firstRun}
          <p class="modal-sub">
            iBinds derives the control profiles folder from your iRacing folder.
          </p>
        {/if}
      </div>
      {#if !firstRun}
        <button class="icon-btn" aria-label="Close settings" onclick={onClose}>
          <Icon name="close" />
        </button>
      {/if}
    </div>

    <div class="modal-body">
      <div class="field">
        <label class="field-label" for="profiles-dir">iRacing folder</label>
        <input
          id="profiles-dir"
          aria-label="iRacing folder"
          data-autofocus
          bind:value={value}
          spellcheck="false"
          autocomplete="off"
        />
        <span class="field-hint">Usually Documents/iRacing.</span>
      </div>

      <div class="field">
        <span class="field-label">Profiles folder</span>
        <code class="derived-path" aria-label="Profiles folder">{derivedProfilesDir}</code>
        <span class="field-hint">Derived from the iRacing folder.</span>
      </div>

      <div class="field">
        <label class="field-label" for="default-device">Default device</label>
        <select
          id="default-device"
          aria-label="Default device"
          value={deviceValue}
          onchange={(e) => (deviceValue = e.currentTarget.value)}
        >
          <option value="" selected={deviceValue === ''}>First available</option>
          {#each devices as d (d.instanceGUID)}
            <option value={d.instanceGUID} selected={d.instanceGUID === deviceValue}>{d.name}</option>
          {/each}
          {#if deviceValue && !devices.some((d) => d.instanceGUID === deviceValue)}
            <option value={deviceValue} selected>Saved device (not detected)</option>
          {/if}
        </select>
        <span class="field-hint">Preselected when you assign a new key or button.</span>
      </div>

      {#if error}<p class="error" role="alert">{error}</p>{/if}
      {#if message}<p class="message">{message}</p>{/if}
    </div>

    <div class="modal-actions">
      {#if !firstRun}<button class="btn" onclick={onClose}>Close</button>{/if}
      <button
        class="btn btn-primary"
        onclick={save}
        disabled={firstRun && !value.trim()}
      >
        {firstRun ? 'Confirm' : 'Save'}
      </button>
    </div>
  </div>
</div>

<style>
  .settings-modal {
    max-width: 34rem;
  }
  .modal-sub {
    margin-top: 0.1rem;
    color: var(--muted);
    font-size: 0.78rem;
  }
  .derived-path {
    display: block;
    overflow-wrap: anywhere;
    padding: 0.45rem 0.55rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    color: var(--muted);
    font-size: 0.78rem;
  }
  .error {
    color: var(--danger);
    font-size: 0.82rem;
  }
  .message {
    color: var(--success);
    font-size: 0.82rem;
    font-weight: 600;
  }
</style>
