<script lang="ts">
  import type { ControlDTO } from '../lib/api';
  import { copyableControl, currentBinding } from '../lib/store';

  let {
    controls,
    selected,
    search,
    onSearch,
    onToggle,
    onToggleAll,
    onEdit,
    onCopySelected,
  }: {
    controls: ControlDTO[];
    selected: Record<string, boolean>;
    search: string;
    onSearch: (v: string) => void;
    onToggle: (name: string) => void;
    onToggleAll: () => void;
    onEdit: (name: string) => void;
    onCopySelected: () => void;
  } = $props();

  const filtered = $derived(
    controls.filter((c) => {
      if (!search) return true;
      const q = search.toLowerCase();
      return c.name.toLowerCase().includes(q) || c.displayName.toLowerCase().includes(q);
    }),
  );
  const selectable = $derived(filtered.filter(copyableControl));
  const allSelected = $derived(
    selectable.length > 0 && selectable.every((c) => selected[c.name]),
  );
  const selectedCount = $derived(selectable.filter((c) => selected[c.name]).length);
</script>

<div class="controls">
  <input
    type="search"
    placeholder="Search controls…"
    value={search}
    oninput={(e) => onSearch(e.currentTarget.value)}
  />

  <table>
    <thead>
      <tr>
        <th>
          <input type="checkbox" checked={allSelected} onchange={onToggleAll} />
        </th>
        <th>Control</th>
        <th>Binding</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {#each filtered as c (c.name)}
        <tr>
          <td>
            <input
              type="checkbox"
              disabled={!copyableControl(c)}
              checked={selected[c.name] ?? false}
              onchange={() => onToggle(c.name)}
            />
          </td>
          <td>{c.displayName}</td>
          <td>{currentBinding(c)}</td>
          <td><button onclick={() => onEdit(c.name)}>Edit</button></td>
        </tr>
      {/each}
    </tbody>
  </table>

  <div class="footer">
    <span>{selectedCount} selected</span>
    <button onclick={onCopySelected}>Copy selected →</button>
  </div>
</div>

<style>
  .controls {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  table {
    border-collapse: collapse;
    width: 100%;
  }
  th,
  td {
    text-align: left;
    padding: 0.25rem 0.5rem;
    border-bottom: 1px solid #eee;
  }
  .footer {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }
</style>
