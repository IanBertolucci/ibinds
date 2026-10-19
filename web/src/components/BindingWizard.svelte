<script lang="ts">
  import { api, keyToByte, type BindingInput, type ControlDTO, type DeviceInfo } from '../lib/api';

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
    function handler(e: KeyboardEvent) {
      e.preventDefault();
      const b = keyToByte(e.key);
      if (b === null) {
        message = 'Only printable A-Z and 0-9 keys are supported (F-keys and arrows are not).';
        return;
      }
      message = '';
      onBinding({ key: String.fromCharCode(b) });
      window.removeEventListener('keydown', handler);
    }
    window.addEventListener('keydown', handler);
  }

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
  <label>
    Control
    <select bind:value={controlName} onchange={(e) => pickControl(e.currentTarget.value)}>
      <option value="">Select a control…</option>
      {#each editable as c (c.name)}
        <option value={c.name}>{c.displayName}</option>
      {/each}
    </select>
  </label>

  <div class="capture">
    <button onclick={captureKey}>Capture key</button>
    <button onclick={captureButton} disabled={capturing}>
      {capturing ? 'Capturing…' : 'Capture button'}
    </button>
  </div>

  <div class="dropdowns">
    <label>
      Device
      <select bind:value={selectedDevice}>
        {#each devices as d (d.instanceGUID)}
          <option value={d.instanceGUID}>{d.name}</option>
        {/each}
      </select>
    </label>
    {#if device}
      <label>
        Button
        <select bind:value={buttonNumber}>
          {#each Array.from({ length: device.buttonCount > 0 ? device.buttonCount : 160 }, (_, i) => i) as n}
            <option value={n}>{n}</option>
          {/each}
        </select>
      </label>
      <button onclick={pickButton}>Use button</button>
    {/if}
    <label>
      Key
      <select bind:value={keyDropdown}>
        {#each 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'.split('') as k}
          <option value={k}>{k}</option>
        {/each}
      </select>
    </label>
    <button onclick={pickKey}>Use key</button>
  </div>

  {#if binding.key}
    <p>Binding: Key {binding.key}</p>
  {:else if binding.button !== undefined}
    <p>Binding: Btn {binding.button}</p>
  {/if}

  {#if message}<p class="message">{message}</p>{/if}

  <button onclick={onPlan} disabled={!controlName || (!binding.key && binding.button === undefined)}>
    Preview
  </button>
</div>

<style>
  .wizard {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .capture,
  .dropdowns {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
  }
  .message {
    color: #b00;
  }
</style>
