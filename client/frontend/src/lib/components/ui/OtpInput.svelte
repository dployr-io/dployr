<script lang="ts">
  import { onMount } from 'svelte';
  import { otp } from '../../../stores';     
  import { get } from 'svelte/store';

  const length = 6;
  let codes: string[] = [];
  let inputs: HTMLInputElement[] = [];

  onMount(() => {
    const initial = get(otp) || '';
    codes = Array.from({ length }, (_, i) => initial.charAt(i));
  });

  $: otp.set(codes.join(''));

  function focus(idx: number) {
    inputs[idx]?.focus();
  }

  function onInput(e: Event, idx: number) {
    const el = e.target as HTMLInputElement;
    const v = el.value.slice(-1);
    codes[idx] = v;
    el.value = v;
    if (v && idx < length - 1) focus(idx + 1);
  }

  function onKeydown(e: KeyboardEvent, idx: number) {
    const el = e.currentTarget as HTMLInputElement;
    if (e.key === 'Backspace' && !el.value && idx > 0) {
      codes[idx - 1] = '';
      focus(idx - 1);
    }
  }

  function onPaste(e: ClipboardEvent) {
    e.preventDefault();
    const text = e.clipboardData?.getData('text') || '';
    text
      .slice(0, length)
      .split('')
      .forEach((ch, i) => {
        codes[i] = ch;
        if (inputs[i]) inputs[i].value = ch;
      });
    focus(Math.min(text.length, length - 1));
  }
</script>

<div class="flex gap-2 justify-center" on:paste={onPaste}>
  {#each Array(length) as _, idx}
    <input
      type="tel"
      inputmode="numeric"
      maxlength="1"
      bind:this={inputs[idx]}
      class="w-12 h-12 text-center text-lg rounded-lg border border-gray-300 focus:border-[#195B5E] outline-none transition"
      value={codes[idx]}
      on:input={(e) => onInput(e, idx)}
      on:keydown={(e) => onKeydown(e, idx)}
    />
  {/each}
</div>
