<script lang="ts">
  import { untrack } from 'svelte';
  import { api, type DeviceInfo, type Settings as SettingsDTO } from '../lib/api';
  import Icon from './Icon.svelte';
  import { focusTrap } from '../lib/focusTrap';

  let {
    profilesDir,
    defaultDeviceInstanceGUID,
    onSaved,
    onClose,
  }: {
    profilesDir: string;
    defaultDeviceInstanceGUID: string;
    onSaved: (settings: SettingsDTO) => void;
    onClose: () => void;
  } = $props();

  let value = $state(untrack(() => profilesDir));
  let deviceValue = $state(untrack(() => defaultDeviceInstanceGUID));
  let devices = $state<DeviceInfo[]>([]);
  let message = $state('');
  let error = $state('');

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
      const res = await api.saveSettings(value, deviceValue);
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
  onclick={(e) => e.target === e.currentTarget && onClose()}
>
  <div
    class="modal settings-modal"
    use:focusTrap={onClose}
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    aria-labelledby="settings-title"
  >
    <div class="modal-header">
      <h2 id="settings-title">Settings</h2>
      <button class="icon-btn" aria-label="Close settings" onclick={onClose}>
        <Icon name="close" />
      </button>
    </div>

    <div class="modal-body">
      <div class="field">
        <label class="field-label" for="profiles-dir">Profiles directory</label>
        <input
          id="profiles-dir"
          aria-label="Profiles directory"
          data-autofocus
          bind:value={value}
          spellcheck="false"
          autocomplete="off"
        />
        <span class="field-hint">The folder that contains your iRacing control profiles.</span>
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
      <button class="btn" onclick={onClose}>Close</button>
      <button class="btn btn-primary" onclick={save}>Save</button>
    </div>
  </div>
</div>

<style>
  .settings-modal {
    max-width: 34rem;
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
