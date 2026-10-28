import type { Action } from 'svelte/action';

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

// focusTrap keeps keyboard focus inside a dialog, closes it on Escape, and
// restores focus to the element that opened it.
export const focusTrap: Action<HTMLElement, () => void> = (node, onClose) => {
  let close = onClose;
  const previous = document.activeElement as HTMLElement | null;

  function focusable(): HTMLElement[] {
    return [...node.querySelectorAll<HTMLElement>(FOCUSABLE)];
  }

  function initial(): HTMLElement | undefined {
    return node.querySelector<HTMLElement>('[data-autofocus]') ?? focusable()[0];
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      close();
      return;
    }
    if (e.key !== 'Tab') return;

    const items = focusable();
    if (items.length === 0) {
      e.preventDefault();
      node.focus();
      return;
    }

    const first = items[0];
    const last = items[items.length - 1];
    const active = document.activeElement as HTMLElement | null;

    if (!node.contains(active)) {
      e.preventDefault();
      first.focus();
      return;
    }
    if (e.shiftKey && (active === first || active === node)) {
      e.preventDefault();
      last.focus();
      return;
    }
    if (!e.shiftKey && active === last) {
      e.preventDefault();
      first.focus();
    }
  }

  node.addEventListener('keydown', onKeydown);
  queueMicrotask(() => {
    (initial() ?? node).focus();
  });

  return {
    update(next: () => void) {
      close = next;
    },
    destroy() {
      node.removeEventListener('keydown', onKeydown);
      if (previous && document.contains(previous)) previous.focus();
    },
  };
};
