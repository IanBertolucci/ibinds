<script lang="ts">
  import { api, keyToByte, type BindingInput, type ControlDTO, type DeviceInfo } from '../lib/api';
  import Icon from './Icon.svelte';

  let {
    controls,
    controlName,
    binding,
    onControl,
    onBinding,
    onPlan,
  }: {
    controls: ControlDTO[];
    controlName: string;
    binding: BindingInput;
    onControl: (name: string) => void;
    onBinding: (b: BindingInput) => void;
    onPlan: () => void;
  } = $props();

  let message = $state('');
  let devices = $state<DeviceInfo[]>([]);
  let selectedDevice = $state('');
  let buttonNumber = $state(0);
  let keyDropdown = $state('U');
  let capturing = $state(false);
  let keyHandler: ((e: KeyboardEvent) => void) | null = null;

  const editable = $derived(
    controls.filter((c) => c.kind === 'unbound' || c.kind === 'key' || c.kind === 'button'),
  );
  const device = $derived(devices.find((d) => d.instanceGUID === selectedDevice));

  $effect(() => {
    api.fetchInputDevices().then((d) => {
      devices = d;
      if (d.length > 0) selectedDevice = d[0].instanceGUID;
    });
  });

  function pickControl(name: string) {
    onControl(name);
    onBinding({});
  }

  function captureKey() {
    if (keyHandler) window.removeEventListener('keydown', keyHandler);
    const handler = (e: KeyboardEvent) => {
      e.preventDefault();
      const b = keyToByte(e.key);
      if (b === null) {
        message = 'Only printable A-Z and 0-9 keys are supported (F-keys and arrows are not).';
        return;
      }
      message = '';
      onBinding({ key: String.fromCharCode(b) });
      window.removeEventListener('keydown', handler);
      keyHandler = null;
    };
    keyHandler = handler;
    window.addEventListener('keydown', handler);
  }

  $effect(() => () => {
    if (keyHandler) window.removeEventListener('keydown', keyHandler);
  });

  async function captureButton() {
    capturing = true;
    try {
      const result = await api.capture(10000);
      if (!result) {
        message = 'No button pressed (timeout).';
      } else {
        onBinding({
          button: result.button,
          deviceInstanceGUID: result.deviceInstanceGUID,
          productGUID: result.productGUID,
        });
      }
    } catch (e) {
      message = String(e);
    } finally {
      capturing = false;
    }
  }

  function pickKey() {
    onBinding({ key: keyDropdown });
  }

  function pickButton() {
    if (!device) return;
    onBinding({
      button: buttonNumber,
      deviceInstanceGUID: device.instanceGUID,
      productGUID: device.productGUID,
    });
  }
</script>

<div class="wizard">
  <label class="field">
    <span class="field-label">Control</span>
    <select data-autofocus bind:value={controlName} onchange={(e) => pickControl(e.currentTarget.value)}>
      <option value="">Select a control…</option>
      {#each editable as c (c.name)}
        <option value={c.name}>{c.displayName}</option>
      {/each}
    </select>
  </label>

  <div class="capture-row">
    <button class="btn" onclick={captureKey}>
      <Icon name="edit" size={15} />
      Capture key
    </button>
    <button class="btn" onclick={captureButton} disabled={capturing}>
      <Icon name="device" size={15} />
      {capturing ? 'Capturing…' : 'Capture button'}
    </button>
  </div>

  <div class="divider"><span>or pick manually</span></div>

  <div class="manual-grid">
    <label class="field">
      <span class="field-label">Device</span>
      <select bind:value={selectedDevice}>
        {#each devices as d (d.instanceGUID)}
          <option value={d.instanceGUID}>{d.name}</option>
        {/each}
      </select>
    </label>

    {#if device}
      <label class="field">
        <span class="field-label">Button</span>
        <select bind:value={buttonNumber}>
          {#each Array.from({ length: device.buttonCount > 0 ? device.buttonCount : 160 }, (_, i) => i) as n}
            <option value={n}>{n}</option>
          {/each}
        </select>
      </label>
      <button class="btn" onclick={pickButton}>Use button</button>
    {/if}

    <label class="field">
      <span class="field-label">Key</span>
      <select bind:value={keyDropdown}>
        {#each 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'.split('') as k}
          <option value={k}>{k}</option>
        {/each}
      </select>
    </label>
    <button class="btn" onclick={pickKey}>Use key</button>
  </div>

  <div class="binding-summary">
    {#if binding.key}
      <span class="field-label">Selected</span>
      <span class="binding-pill mono">Key <span class="key-chip">{binding.key}</span></span>
    {:else if binding.button !== undefined}
      <span class="field-label">Selected</span>
      <span class="binding-pill mono">Button <span class="button-chip">{binding.button}</span></span>
    {:else}
      <span class="muted">No key or button selected yet.</span>
    {/if}
  </div>

  {#if message}<p class="message">{message}</p>{/if}

  <button
    class="btn btn-primary preview"
    onclick={onPlan}
    disabled={!controlName || (!binding.key && binding.button === undefined)}
  >
    Preview
  </button>
</div>

<style>
  .wizard {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
  }
  .capture-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .divider {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    color: var(--faint);
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
  .divider::before,
  .divider::after {
    content: "";
    flex: 1;
    height: 1px;
    background: var(--border);
  }
  .manual-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    align-items: end;
  }
  .manual-grid .field {
    flex: 1 1 8rem;
    min-width: 0;
  }
  .manual-grid .btn {
    flex: 0 0 auto;
  }
  .binding-summary {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
    padding: 0.6rem 0.7rem;
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    font-size: 0.82rem;
  }
  .binding-pill {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.8rem;
    font-weight: 700;
  }
  .button-chip,
  .key-chip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.4rem;
    padding: 0.08rem 0.35rem;
    border: 1px solid var(--border-strong);
    border-radius: 5px;
    background: var(--surface);
    font-size: 0.76rem;
    font-weight: 700;
  }
  .message {
    color: var(--danger);
    font-size: 0.8rem;
  }
  .preview {
    align-self: flex-start;
  }
  @media (max-width: 640px) {
    .manual-grid .field,
    .manual-grid .btn {
      flex-basis: 100%;
    }
  }
</style>
