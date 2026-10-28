<script lang="ts">
  import type { ControlDTO } from '../lib/api';
  import { copyableControl } from '../lib/store';
  import Icon from './Icon.svelte';

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
    onToggleAll: (names: string[]) => void;
    onEdit: (name: string) => void;
    onCopySelected: () => void;
  } = $props();

  type Filter = 'all' | 'bound' | 'unbound' | 'copyable';
  let filter = $state<Filter>('all');

  const filters: { id: Filter; label: string }[] = [
    { id: 'all', label: 'All' },
    { id: 'bound', label: 'Bound' },
    { id: 'unbound', label: 'Unbound' },
    { id: 'copyable', label: 'Copyable' },
  ];

  function isBound(c: ControlDTO): boolean {
    return c.kind !== 'unbound';
  }

  function buttonList(c: ControlDTO): number[] {
    return c.buttons && c.buttons.length > 0 ? c.buttons : [c.button];
  }

  const filtered = $derived(
    controls.filter((c) => {
      if (filter === 'bound' && !isBound(c)) return false;
      if (filter === 'unbound' && c.kind !== 'unbound') return false;
      if (filter === 'copyable' && !copyableControl(c)) return false;
      if (!search) return true;
      const q = search.toLowerCase();
      return c.name.toLowerCase().includes(q) || c.displayName.toLowerCase().includes(q);
    }),
  );
  const selectable = $derived(filtered.filter(copyableControl));
  const allSelected = $derived(
    selectable.length > 0 && selectable.every((c) => selected[c.name]),
  );
  const selectedCount = $derived(
    controls.filter((c) => copyableControl(c) && selected[c.name]).length,
  );
  const visibleSelectedCount = $derived(
    selectable.filter((c) => selected[c.name]).length,
  );
  const hiddenSelectedCount = $derived(selectedCount - visibleSelectedCount);
</script>

<div class="controls">
  <div class="toolbar">
    <label class="control-search">
      <Icon name="search" size={15} />
      <input
        type="search"
        placeholder="Search controls…"
        aria-label="Search controls"
        value={search}
        oninput={(e) => onSearch(e.currentTarget.value)}
      />
    </label>

    <div class="chips" role="group" aria-label="Filter controls">
      {#each filters as f (f.id)}
        <button class="chip" class:active={filter === f.id} onclick={() => (filter = f.id)}>
          {f.label}
        </button>
      {/each}
    </div>

    <span class="result-count">{filtered.length} of {controls.length}</span>
  </div>

  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th class="check">
            <input
              type="checkbox"
              checked={allSelected}
              disabled={selectable.length === 0}
              onchange={() => onToggleAll(selectable.map((c) => c.name))}
              aria-label="Select all copyable controls"
            />
          </th>
          <th>Control</th>
          <th>Binding</th>
          <th class="actions">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each filtered as c (c.name)}
          <tr class:selected={selected[c.name]}>
            <td class="check">
              <input
                type="checkbox"
                disabled={!copyableControl(c)}
                checked={selected[c.name] ?? false}
                onchange={() => onToggle(c.name)}
                aria-label={`Select ${c.displayName}`}
              />
            </td>
            <td class="control-cell" title={c.name}>{c.displayName}</td>
            <td class="binding-cell">
              <div class="binding-content">
                {#if c.kind === 'key'}
                  <span class="binding mono">
                    <span class="binding-kind">Key</span>{' '}
                    <span class="key-chip">{c.key}</span>
                  </span>
                {:else if c.kind === 'button'}
                  <span class="binding mono">
                    <span class="binding-kind">Btn</span>{' '}
                    {#each buttonList(c) as b, i (i)}
                      {#if i > 0}{' + '}{/if}
                      <span class="button-chip">{b}</span>
                    {/each}
                  </span>
                  {#if c.deviceName}<span class="device-name">{c.deviceName}</span>{/if}
                {:else if c.kind === 'axis'}
                  <span class="binding muted">Axis</span>
                {:else}
                  <span class="binding muted">None</span>
                {/if}
              </div>
            </td>
            <td class="actions">
              <button class="btn btn-ghost row-action" onclick={() => onEdit(c.name)}>
                <Icon name="edit" size={15} />
                Edit
              </button>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="4" class="empty-row">No controls match your search.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <div class="footer">
    <span class="selection-count">
      {selectedCount} selected
      {#if hiddenSelectedCount > 0}<span class="muted">({hiddenSelectedCount} hidden)</span>{/if}
    </span>
    <button class="btn" onclick={onCopySelected}>
      <Icon name="copy" size={15} />
      Copy selected →
    </button>
  </div>
</div>

<style>
  .controls {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.6rem;
  }
  .control-search {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    flex: 1 1 15rem;
    min-width: 12rem;
    padding: 0 0.6rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface);
    color: var(--faint);
  }
  .control-search:focus-within {
    border-color: var(--accent);
  }
  .control-search input {
    width: 100%;
    min-width: 0;
    border: 0;
    padding: 0.5rem 0;
    background: transparent;
    outline: none;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    padding: 0.2rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--surface);
  }
  .chip {
    padding: 0.32rem 0.65rem;
    border: 0;
    border-radius: 999px;
    background: transparent;
    color: var(--muted);
    font-size: 0.78rem;
    font-weight: 600;
  }
  .chip:hover {
    background: var(--surface-2);
    color: var(--text);
  }
  .chip.active {
    background: var(--accent-soft);
    color: var(--accent-ink);
  }
  .result-count {
    margin-left: auto;
    color: var(--faint);
    font-size: 0.76rem;
    font-variant-numeric: tabular-nums;
  }
  .table-wrap {
    max-height: calc(100vh - 270px);
    overflow: auto;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
    box-shadow: var(--shadow-sm);
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.84rem;
  }
  th {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 0.55rem 0.7rem;
    border-bottom: 1px solid var(--border);
    background: var(--surface-2);
    color: var(--muted);
    font-size: 0.7rem;
    font-weight: 800;
    letter-spacing: 0.05em;
    text-align: left;
    text-transform: uppercase;
    white-space: nowrap;
  }
  td {
    padding: 0.5rem 0.7rem;
    border-bottom: 1px solid var(--border);
    vertical-align: middle;
  }
  tbody tr:last-child td {
    border-bottom: 0;
  }
  tbody tr:hover {
    background: #fafcfd;
  }
  tbody tr.selected {
    background: var(--accent-soft);
  }
  .check {
    width: 2.2rem;
    text-align: center;
  }
  .check input {
    accent-color: var(--accent);
  }
  .control-cell {
    font-weight: 600;
    white-space: nowrap;
  }
  .binding-content {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.35rem;
  }
  .binding {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    color: var(--text);
    font-size: 0.8rem;
    white-space: nowrap;
  }
  .binding-kind {
    color: var(--muted);
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
    background: var(--surface-2);
    font-size: 0.76rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .device-name {
    color: var(--faint);
    font-size: 0.74rem;
  }
  .actions {
    width: 1%;
    text-align: right;
    white-space: nowrap;
  }
  .row-action {
    min-height: 28px;
    padding: 0.28rem 0.5rem;
    color: var(--muted);
  }
  .row-action:hover {
    color: var(--accent-ink);
  }
  .empty-row {
    padding: 2rem 1rem;
    color: var(--muted);
    text-align: center;
  }
  .footer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .selection-count {
    color: var(--muted);
    font-size: 0.8rem;
    font-variant-numeric: tabular-nums;
  }
  @media (max-width: 1100px) {
    .table-wrap {
      max-height: 60vh;
    }
  }
  @media (max-width: 640px) {
    .result-count {
      margin-left: 0;
    }
    .control-cell {
      white-space: normal;
    }
  }
</style>
