import { mount } from 'svelte';
import App from './App.svelte';
import { token } from './lib/api';

token.init();

export default mount(App, {
  target: document.getElementById('app')!,
});
