<script lang="ts">
  import { EventsOn } from '../../../../wailsjs/runtime/runtime'
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { wsconsole } from '../../../../src/stores';
  import { 
    ConnectSsh, 
    StartTerminalWebSocket, 
    SendTerminalInput, 
    ResizeTerminal, 
    DisconnectTerminal
  } from '../../../../wailsjs/go/main/App';

  let hostname = '';
  let port = 22;
  let username = 'techcyborg';
  let password = '';
  let status = 'disconnected';
  let unsubscribeFunctions: Array<() => void> = [];

  const themes = {
    dark: {
      background: '#000000',
      foreground: '#ffffff',
      cursor: '#ffffff',
      selection: '#ffffff40'
    }
  };

  onMount(() => {
    initializeTerminal();
    setupEventListeners();
  });

  onDestroy(() => {
    cleanup();
  });

  function initializeTerminal() {
    $wsconsole.terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
      theme: themes.dark,
      cols: 80,
      rows: Math.floor((window.innerHeight - 300) / 20), // Dynamic rows
      scrollback: 10000,
      scrollOnUserInput: true,
      convertEol: true,
      disableStdin: false,
      cursorStyle: 'block'
    });

    $wsconsole.fitAddon = new FitAddon();
    $wsconsole.terminal.loadAddon($wsconsole.fitAddon);
    $wsconsole.terminal.open($wsconsole.terminalElement);
    $wsconsole.fitAddon.fit();

    // Handle terminal input
    $wsconsole.terminal.onData((data: string) => {
      if (status === 'connected') {
        SendTerminalInput(data);
      }
    });

    // Handle window resize
    window.addEventListener('resize', () => {
      if ($wsconsole.fitAddon) {
        $wsconsole.fitAddon.fit();
        if (status === 'connected') {
          ResizeTerminal($wsconsole.terminal.cols, $wsconsole.terminal.rows);
        }
      }
    });

    $wsconsole.terminal.writeln('SSH Terminal Ready');
    $wsconsole.terminal.write('Enter connection details and click Connect.\r\n');
  }

  function setupEventListeners() {
    const unsubConnected = EventsOn("terminal:connected", () => {
      status = 'connected';
      $wsconsole.statusMessage = `Connected to ${username}@${hostname}:${port}`;
      $wsconsole.terminal.clear();
      $wsconsole.terminal.write('\x1b[H\x1b[2J'); // Clear screen and move cursor to home
      $wsconsole.terminal.focus();
      $wsconsole.fitAddon.fit();
      ResizeTerminal($wsconsole.terminal.cols, $wsconsole.terminal.rows);
    });

    const unsubDisconnected = EventsOn("terminal:disconnected", () => {
      status = 'disconnected';
      $wsconsole.statusMessage = 'Disconnected';
      $wsconsole.terminal.writeln('\r\nConnection closed');
    });

    const unsubOutput = EventsOn("terminal:output", (data: string) => {
      $wsconsole.terminal.write(data);
    });

    unsubscribeFunctions = [unsubConnected, unsubDisconnected, unsubOutput];
  }

  async function connect() {
    if (!hostname || !username || !password) {
      $wsconsole.errorMessage = 'Fill all fields';
      return;
    }

    status = 'connecting';
    $wsconsole.statusMessage = 'Connecting...';
    $wsconsole.errorMessage = '';

    try {
      const result = await ConnectSsh(hostname, port, username, password);
      await StartTerminalWebSocket(hostname, result.sessionId);
    } catch (error) {
      status = 'disconnected';
      $wsconsole.errorMessage = `Connection failed: ${error}`;
      $wsconsole.statusMessage = 'Connection failed';
    }
  }

  async function disconnect() {
    await DisconnectTerminal();
  }

  function clearTerminal() {
    $wsconsole.terminal.clear();
  }

  function cleanup() {
    unsubscribeFunctions.forEach(unsub => unsub());
    DisconnectTerminal();
    if ($wsconsole.terminal) {
      $wsconsole.terminal.dispose();
    }
  }
</script>

<div class="ssh-terminal">
  <div class="header">
    <div class="connection-form">
      <div class="form-group">
        <label>Username</label>
        <input bind:value={username} placeholder="user" />
      </div>
      <div class="form-group">
        <label>Password</label>
        <input type="password" bind:value={password} placeholder="password" />
      </div>
      <button class="connect-btn" on:click={connect} disabled={status === 'connecting'}>
        {status === 'connecting' ? 'Connecting...' : 'Connect'}
      </button>
    </div>
  </div>

  <div class="status {status}">
    {#if $wsconsole.errorMessage}
      <p>{$wsconsole.errorMessage}</p>
    {:else}
      <p>{$wsconsole.statusMessage}</p>
    {/if}
  </div>

  <div class="terminal-container">
    <div class="terminal-controls">
      <button on:click={clearTerminal} disabled={status !== 'connected'}>Clear</button>
      <button on:click={disconnect} disabled={status !== 'connected'}>Disconnect</button>
    </div>
    <div class="terminal-wrapper">
      <div bind:this={$wsconsole.terminalElement} class="terminal"></div>
    </div>
  </div>
</div>

<style>
  .ssh-terminal {
    height: 100vh;
    display: flex;
    flex-direction: column;
    font-family: 'Segoe UI', sans-serif;
    background: #1e1e1e;
    color: #ffffff;
  }

  .header {
    background: #2d2d30;
    padding: 1rem;
    border-bottom: 1px solid #3e3e42;
  }

  .connection-form {
    display: flex;
    gap: 1rem;
    align-items: center;
    flex-wrap: wrap;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .form-group label {
    font-size: 0.875rem;
    color: #cccccc;
  }

  .form-group input {
    padding: 0.5rem;
    border: 1px solid #3e3e42;
    border-radius: 4px;
    background: #1e1e1e;
    color: #ffffff;
    font-size: 0.875rem;
    min-width: 120px;
  }

  .connect-btn {
    background: #0e639c;
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 4px;
    cursor: pointer;
    margin-top: auto;
  }

  .status {
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
  }

  .status.connected { background: #0e5a0e; color: #4caf50; }
  .status.connecting { background: #663c00; color: #ff9800; }
  .status.disconnected { background: #5a0e0e; color: #f44336; }

  .error-message {
    background: #5a0e0e;
    color: #f44336;
    padding: 0.75rem;
    margin: 0.5rem;
    border-radius: 4px;
    max-height: 32px;           /* Limit error message height */
    align-self: center;
  }

  .terminal-container {
    flex: 1;
    padding: 1rem;
    background: #0c0c0c;
    display: flex;
    flex-direction: column;
    min-height: 0;  /* Allow flex shrinking */
    height: calc(100vh - 250px);  /* Fixed height with room for errors */
    overflow: hidden;             /* Prevent container overflow */
  }

  .terminal-controls {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .terminal-controls button {
    background: #3c3c3c;
    color: #ffffff;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.75rem;
  }

  .terminal-wrapper {
    flex: 1;
    background: #000000;
    border-radius: 4px;
    padding: 0.5rem;
    overflow: auto;
    text-align: left;  /* Force left alignment */
    display: flex;
    flex-direction: column;
  }

  .terminal {
    height: 100%;
    width: 100%;
    text-align: left;  /* Ensure terminal content is left-aligned */
  }

  /* Override any inherited text alignment */
  .terminal * {
    text-align: left !important;
  }
</style>