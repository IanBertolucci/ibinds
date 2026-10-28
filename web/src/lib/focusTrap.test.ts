import { describe, test, expect, vi } from 'vitest';
import { focusTrap } from './focusTrap';

function press(node: HTMLElement, key: string, shiftKey = false) {
  node.dispatchEvent(new KeyboardEvent('keydown', { key, shiftKey, bubbles: true, cancelable: true }));
}

async function settle() {
  await new Promise((r) => setTimeout(r, 0));
}

function makeDialog() {
  const dialog = document.createElement('div');
  dialog.tabIndex = -1;
  dialog.innerHTML = '<button id="first">first</button><button id="last">last</button>';
  document.body.appendChild(dialog);
  return dialog;
}

describe('focusTrap', () => {
  test('focuses the first control and keeps Tab inside', async () => {
    const outside = document.createElement('button');
    document.body.appendChild(outside);
    outside.focus();
    const dialog = makeDialog();
    const onClose = vi.fn();

    const action = focusTrap(dialog, onClose);
    await settle();
    expect(document.activeElement).toBe(dialog.querySelector('#first'));

    (dialog.querySelector('#last') as HTMLElement).focus();
    press(dialog, 'Tab');
    expect(document.activeElement).toBe(dialog.querySelector('#first'));

    press(dialog, 'Tab', true);
    expect(document.activeElement).toBe(dialog.querySelector('#last'));

    press(dialog, 'Escape');
    expect(onClose).toHaveBeenCalledTimes(1);

    action.destroy();
    expect(document.activeElement).toBe(outside);

    outside.remove();
    dialog.remove();
  });

  test('focuses the container when there is no focusable child', async () => {
    const dialog = document.createElement('div');
    dialog.tabIndex = -1;
    document.body.appendChild(dialog);
    const action = focusTrap(dialog, vi.fn());
    await settle();
    expect(document.activeElement).toBe(dialog);
    press(dialog, 'Tab');
    expect(document.activeElement).toBe(dialog);
    action.destroy();
    dialog.remove();
  });
});
