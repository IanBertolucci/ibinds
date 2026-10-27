import { describe, test, expect, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import Sidebar from './Sidebar.svelte';
import type { ProfileDTO } from '../lib/api';

const profiles: ProfileDTO[] = [
  { id: 'Baseline', displayName: 'Baseline', isBaseline: true, hasControls: true, hasCalib: false },
  { id: 'f4', displayName: 'F4', isBaseline: false, hasControls: true, hasCalib: false },
  { id: 'f5', displayName: 'F5', isBaseline: false, hasControls: true, hasCalib: false },
];

function mountSidebar(props: Partial<Parameters<typeof Sidebar>[0]> = {}) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onSelect = vi.fn();
  const app = mount(Sidebar, {
    target,
    props: { profiles, activeId: 'Baseline', onSelect, ...props },
  });
  return { target, app, onSelect };
}

describe('Sidebar', () => {
  test('renders one item per profile with a baseline badge', () => {
    const { target, app } = mountSidebar();
    const items = [...target.querySelectorAll('button')].map((b) => b.textContent?.trim());
    expect(items).toHaveLength(3);
    expect(items[0]).toContain('Baseline');
    expect(items[0]).toContain('(baseline)');
    unmount(app);
    target.remove();
  });

  test('clicking a profile calls onSelect', () => {
    const { target, app, onSelect } = mountSidebar();
    const f4 = [...target.querySelectorAll('button')].find((b) => b.textContent?.includes('F4'));
    (f4 as HTMLButtonElement).click();
    expect(onSelect).toHaveBeenCalledWith('f4');
    unmount(app);
    target.remove();
  });

  test('active profile is highlighted', () => {
    const { target, app } = mountSidebar({ activeId: 'f4' });
    const f4 = [...target.querySelectorAll('button')].find((b) => b.textContent?.includes('F4'));
    expect(f4?.className).toContain('active');
    unmount(app);
    target.remove();
  });
});
