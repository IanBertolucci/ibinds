import { describe, test, expect, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import Settings from './Settings.svelte';

async function settle() {
  for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    await tick();
  }
}

function mountSettings() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onSaved = vi.fn();
  const onClose = vi.fn();
  const app = mount(Settings, {
    target,
    props: { profilesDir: 'C:\\Users\\Ian\\Documents\\iRacing', onSaved, onClose },
  });
  return { target, app, onSaved, onClose };
}

describe('Settings', () => {
  test('shows the profiles directory field', () => {
    const { target, app } = mountSettings();
    const input = target.querySelector('input[aria-label="Profiles directory"]') as HTMLInputElement;
    expect(input).toBeTruthy();
    expect(input.value).toBe('C:\\Users\\Ian\\Documents\\iRacing');
    unmount(app);
    target.remove();
  });

  test('save failure shows an error instead of a success message', async () => {
    sessionStorage.setItem('token', 'abc');
    globalThis.fetch = vi.fn(async () => ({ ok: false, status: 400, json: async () => ({}) } as Response)) as typeof fetch;
    const { target, app } = mountSettings();
    const save = [...target.querySelectorAll('button')].find(
      (b) => b.textContent?.trim() === 'Save',
    ) as HTMLButtonElement;
    save.click();
    await settle();
    expect(target.querySelector('[role="alert"]')).toBeTruthy();
    expect(target.textContent).not.toContain('Saved.');
    unmount(app);
    target.remove();
  });

  test('close button calls onClose', () => {
    const { target, app, onClose } = mountSettings();
    const close = target.querySelector('button[aria-label="Close settings"]') as HTMLButtonElement;
    close.click();
    expect(onClose).toHaveBeenCalled();
    unmount(app);
    target.remove();
  });
});
