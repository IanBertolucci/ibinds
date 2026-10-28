import { describe, test, expect, vi, beforeEach } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import Settings from './Settings.svelte';

async function settle() {
  for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    await tick();
  }
}

function jsonResponse(v: unknown) {
  return Promise.resolve({ ok: true, status: 200, json: async () => v } as Response);
}

function mountSettings(overrides: Record<string, unknown> = {}) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onSaved = vi.fn();
  const onClose = vi.fn();
  const app = mount(Settings, {
    target,
    props: {
      profilesDir: 'C:\\Users\\Ian\\Documents\\iRacing',
      defaultDeviceInstanceGUID: '',
      onSaved,
      onClose,
      ...overrides,
    },
  });
  return { target, app, onSaved, onClose };
}

describe('Settings', () => {
  beforeEach(() => {
    sessionStorage.setItem('token', 'abc');
    globalThis.fetch = vi.fn(async () => jsonResponse([])) as typeof fetch;
  });

  test('shows the profiles directory field', () => {
    const { target, app } = mountSettings();
    const input = target.querySelector('input[aria-label="Profiles directory"]') as HTMLInputElement;
    expect(input).toBeTruthy();
    expect(input.value).toBe('C:\\Users\\Ian\\Documents\\iRacing');
    unmount(app);
    target.remove();
  });

  test('shows the default device field', () => {
    const { target, app } = mountSettings({ defaultDeviceInstanceGUID: '{BBB}' });
    const select = target.querySelector('select[aria-label="Default device"]') as HTMLSelectElement;
    expect(select).toBeTruthy();
    expect(select.value).toBe('{BBB}');
    unmount(app);
    target.remove();
  });

  test('save sends the chosen default device', async () => {
    const devices = [
      { name: 'Alpha Wheel', instanceGUID: '{AAA}', productGUID: 'p1', buttonCount: 16 },
      { name: 'Beta Wheel', instanceGUID: '{BBB}', productGUID: 'p2', buttonCount: 16 },
    ];
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/api/input/devices')) return jsonResponse(devices);
      if (url.includes('/api/settings')) {
        return jsonResponse({ profilesDir: 'D', defaultDeviceInstanceGUID: '{BBB}' });
      }
      return jsonResponse({});
    }) as typeof fetch;

    const { target, app } = mountSettings();
    await settle();

    const select = target.querySelector('select[aria-label="Default device"]') as HTMLSelectElement;
    select.value = '{BBB}';
    select.dispatchEvent(new Event('change', { bubbles: true }));
    await settle();

    const save = [...target.querySelectorAll('button')].find(
      (b) => b.textContent?.trim() === 'Save',
    ) as HTMLButtonElement;
    save.click();
    await settle();

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    const saveCall = calls.find((c) => String(c[0]).includes('/api/settings'));
    expect(saveCall).toBeTruthy();
    const body = JSON.parse(saveCall![1]!.body as string);
    expect(body).toEqual({
      profilesDir: 'C:\\Users\\Ian\\Documents\\iRacing',
      defaultDeviceInstanceGUID: '{BBB}',
    });
    unmount(app);
    target.remove();
  });

  test('save failure shows an error instead of a success message', async () => {
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
