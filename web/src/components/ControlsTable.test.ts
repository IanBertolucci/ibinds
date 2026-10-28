import { describe, test, expect, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import ControlsTable from './ControlsTable.svelte';
import type { ControlDTO } from '../lib/api';

function control(partial: Partial<ControlDTO>): ControlDTO {
  return {
    name: '',
    displayName: '',
    category: '',
    subcategory: '',
    kind: 'unbound',
    key: '',
    button: 0,
    buttons: [],
    deviceName: '',
    axisName: '',
    ...partial,
  };
}

const controls: ControlDTO[] = [
  control({ name: 'A', displayName: 'Alpha', category: 'Cockpit', kind: 'key', key: 'U' }),
  control({ name: 'B', displayName: 'Beta', category: 'Cockpit', kind: 'button', button: 131, buttons: [131] }),
  control({ name: 'C', displayName: 'Gamma', category: 'Cockpit', kind: 'button', button: 7, buttons: [7, 33] }),
  control({ name: 'D', displayName: 'Delta', category: "Driver's View", subcategory: 'Custom Sectors', kind: 'axis' }),
  control({ name: 'E', displayName: 'Epsilon', category: "Driver's View", subcategory: 'Active Reset', kind: 'unbound' }),
];

function mountTable(props: Partial<Parameters<typeof ControlsTable>[0]> = {}) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onEdit = vi.fn();
  const onCopySelected = vi.fn();
  const onToggle = vi.fn();
  const onToggleAll = vi.fn();
  const app = mount(ControlsTable, {
    target,
    props: {
      controls,
      selected: {},
      search: '',
      onSearch: vi.fn(),
      onToggle,
      onToggleAll,
      onEdit,
      onCopySelected,
      ...props,
    },
  });
  return { target, app, onEdit, onCopySelected, onToggle, onToggleAll };
}

describe('ControlsTable', () => {
  test('enables checkbox for key and button (single + multi), disables axis', () => {
    const { target, app } = mountTable();
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    const rowByName = (n: string) => {
      const row = rows.find((r) => (r.children[1]?.textContent ?? '') === n);
      return row?.children[0]?.querySelector('input') as HTMLInputElement;
    };
    expect(rowByName('Alpha').disabled).toBe(false);
    expect(rowByName('Beta').disabled).toBe(false);
    expect(rowByName('Gamma').disabled).toBe(false);
    expect(rowByName('Delta').disabled).toBe(true);
    unmount(app);
    target.remove();
  });

  test('renders category and subcategory dividers', () => {
    const { target, app } = mountTable();
    const categories = [...target.querySelectorAll('tbody tr.category-row')].map((r) =>
      r.textContent?.trim(),
    );
    const subcategories = [...target.querySelectorAll('tbody tr.subcategory-row')].map((r) =>
      r.textContent?.trim(),
    );
    expect(categories).toEqual(['Cockpit', "Driver's View"]);
    expect(subcategories).toEqual(['Custom Sectors', 'Active Reset']);
    unmount(app);
    target.remove();
  });

  test('search filters rows', async () => {
    const { target, app } = mountTable({ search: 'beta' });
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    expect(rows).toHaveLength(1);
    unmount(app);
    target.remove();
  });

  test('search also matches the technical control id', () => {
    const { target, app } = mountTable({
      controls: [control({ name: 'TechOnly', displayName: 'Visible Name', category: 'Cockpit' })],
      search: 'techonly',
    });
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    expect(rows).toHaveLength(1);
    expect(rows[0].textContent).toContain('Visible Name');
    unmount(app);
    target.remove();
  });

  test('search matches a category name', () => {
    const { target, app } = mountTable({ search: 'cockpit' });
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    expect(rows).toHaveLength(3);
    expect(rows.map((r) => r.textContent?.trim())).toEqual([
      expect.stringContaining('Alpha'),
      expect.stringContaining('Beta'),
      expect.stringContaining('Gamma'),
    ]);
    unmount(app);
    target.remove();
  });

  test('search matches a subcategory name', () => {
    const { target, app } = mountTable({ search: 'sectors' });
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    expect(rows).toHaveLength(1);
    expect(rows[0].textContent).toContain('Delta');
    unmount(app);
    target.remove();
  });

  test('filter chips narrow the table', async () => {
    const { target, app } = mountTable();
    const unbound = [...target.querySelectorAll('button')].find(
      (b) => b.textContent?.trim() === 'Unbound',
    ) as HTMLButtonElement;
    unbound.click();
    await tick();
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    expect(rows).toHaveLength(1);
    expect(rows[0].textContent).toContain('Epsilon');
    unmount(app);
    target.remove();
  });

  test('select all only toggles visible copyable rows', () => {
    const { target, app, onToggleAll } = mountTable();
    const header = target.querySelector('thead input[type=checkbox]') as HTMLInputElement;
    header.click();
    expect(onToggleAll).toHaveBeenCalledWith(['A', 'B', 'C']);
    unmount(app);
    target.remove();
  });

  test('select all is disabled when the filter has no copyable rows', () => {
    const { target, app, onToggleAll } = mountTable({ search: 'Delta' });
    const header = target.querySelector('thead input[type=checkbox]') as HTMLInputElement;
    expect(header.disabled).toBe(true);
    header.click();
    expect(onToggleAll).not.toHaveBeenCalled();
    unmount(app);
    target.remove();
  });

  test('selection count reports hidden selections', () => {
    const { target, app } = mountTable({ selected: { A: true }, search: 'Beta' });
    expect(target.textContent).toContain('1 selected');
    expect(target.textContent).toContain('(1 hidden)');
    unmount(app);
    target.remove();
  });

  test('binding cell keeps the table cell and wraps flex content inside it', () => {
    const { target, app } = mountTable();
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    const gamma = rows.find((r) => r.textContent?.includes('Gamma'))!;
    const cell = gamma.children[2] as HTMLElement;
    expect(cell.classList.contains('binding-cell')).toBe(true);
    expect(cell.querySelector('.binding-content')).toBeTruthy();
    unmount(app);
    target.remove();
  });

  test('multi-button bindings render one chip per button', () => {
    const { target, app } = mountTable();
    const rows = [...target.querySelectorAll('tbody tr.control-row')];
    const gamma = rows.find((r) => r.textContent?.includes('Gamma'))!;
    const chips = [...gamma.querySelectorAll('.button-chip')].map((c) => c.textContent?.trim());
    expect(chips).toEqual(['7', '33']);
    unmount(app);
    target.remove();
  });

  test('edit button calls onEdit with control name', () => {
    const { target, app, onEdit } = mountTable();
    const editButtons = [...target.querySelectorAll('button')].filter((b) =>
      b.textContent?.trim().startsWith('Edit'),
    );
    editButtons[0].click();
    expect(onEdit).toHaveBeenCalledWith('A');
    unmount(app);
    target.remove();
  });

  test('copy selected button calls onCopySelected', () => {
    const { target, app, onCopySelected } = mountTable({ selected: { A: true } });
    const copyBtn = [...target.querySelectorAll('button')].find((b) =>
      b.textContent?.trim().includes('Copy selected'),
    );
    (copyBtn as HTMLButtonElement).click();
    expect(onCopySelected).toHaveBeenCalled();
    unmount(app);
    target.remove();
  });

  test('renders current binding text', () => {
    const { target, app } = mountTable();
    const text = (target.textContent ?? '').replace(/\s+/g, ' ');
    expect(text).toContain('Key U');
    expect(text).toContain('Btn 7 + 33');
    unmount(app);
    target.remove();
  });
});
