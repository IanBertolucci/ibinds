<script lang="ts">
  import { api } from '../lib/api';

  let { profilesDir, onSaved }: { profilesDir: string; onSaved: (dir: string) => void } = $props();

  let value = $state(profilesDir);
  let message = $state('');

  async function save() {
    try {
      const res = await api.saveSettings(value);
      message = 'Saved.';
      onSaved(res.profilesDir);
    } catch (e) {
      message = String(e);
    }
  }
</script>

<div class="settings">
  <label>
    Profiles directory
    <input bind:value={value} />
  </label>
  <button onclick={save}>Save</button>
  {#if message}<p>{message}</p>{/if}
</div>

<style>
  .settings {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }
  input {
    min-width: 30rem;
  }
</style>
