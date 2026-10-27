import { describe, test, expect, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import ControlsTable from './ControlsTable.svelte';
import type { ControlDTO } from '../lib/api';

function control(partial: Partial<ControlDTO>): ControlDTO {
  return {
    name: '',
    displayName: '',
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
  control({ name: 'A', displayName: 'Alpha', kind: 'key', key: 'U' }),
  control({ name: 'B', displayName: 'Beta', kind: 'button', button: 131, buttons: [131] }),
  control({ name: 'C', displayName: 'Gamma', kind: 'button', button: 7, buttons: [7, 33] }),
  control({ name: 'D', displayName: 'Delta', kind: 'axis' }),
];

function mountTable(props: Partial<Parameters<typeof ControlsTable>[0]> = {}) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onEdit = vi.fn();
  const onCopySelected = vi.fn();
  const onToggle = vi.fn();
  const app = mount(ControlsTable, {
    target,
    props: {
      controls,
      selected: {},
      search: '',
      onSearch: vi.fn(),
      onToggle,
      onToggleAll: vi.fn(),
      onEdit,
      onCopySelected,
      ...props,
    },
  });
  return { target, app, onEdit, onCopySelected, onToggle };
}

describe('ControlsTable', () => {
  test('enables checkbox for key and button (single + multi), disables axis', () => {
    const { target, app } = mountTable();
    const rows = [...target.querySelectorAll('tbody tr')];
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

  test('search filters rows', async () => {
    const { target, app } = mountTable({ search: 'beta' });
    const rows = [...target.querySelectorAll('tbody tr')];
    expect(rows).toHaveLength(1);
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
    const text = target.textContent ?? '';
    expect(text).toContain('Key U');
    expect(text).toContain('Btn 7 + 33');
    unmount(app);
    target.remove();
  });
});
