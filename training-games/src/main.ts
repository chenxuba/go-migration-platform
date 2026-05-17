import Phaser from 'phaser';
import './style.css';
import { ColorMatchScene } from './games/color-shape-match/ColorMatchScene';
import { FireflyTraceScene } from './games/firefly-path-trace/FireflyTraceScene';
import { ShadowTheaterScene } from './games/shadow-theater/ShadowTheaterScene';
import { readLaunchParams } from './platform/hostBridge';

declare global {
  interface Window {
    __COLOR_MATCH_READY__?: boolean;
    __COLOR_MATCH_SCRIPT_LOADED__?: boolean;
    __COLOR_MATCH_STOP_AUDIO__?: () => void;
  }
}

window.__COLOR_MATCH_SCRIPT_LOADED__ = true;

window.addEventListener('error', (event) => {
  showBootError(event.message || '游戏脚本运行失败');
});

window.addEventListener('unhandledrejection', (event) => {
  showBootError(String(event.reason || '游戏脚本运行失败'));
});

window.addEventListener('color-match-ready', hideLoading);

const launchParams = readLaunchParams();
applyGameBootCopy(launchParams.gameId);
const scene =
  launchParams.gameId === 'shadow-theater'
    ? new ShadowTheaterScene(launchParams)
    : launchParams.gameId === 'firefly-path-trace'
    ? new FireflyTraceScene(launchParams)
    : new ColorMatchScene(launchParams);

try {
  new Phaser.Game({
    type: Phaser.CANVAS,
    parent: 'game-root',
    backgroundColor: '#7bdff2',
    scale: {
      mode: Phaser.Scale.FIT,
      autoCenter: Phaser.Scale.CENTER_BOTH,
      width: 1280,
      height: 720,
    },
    dom: {
      createContainer: false,
    },
    input: {
      activePointers: 3,
    },
    scene: [scene],
  });
} catch (error) {
  showBootError(error instanceof Error ? error.message : String(error));
}

function hideLoading(): void {
  const loading = document.querySelector('#game-loading');
  if (!loading) {
    return;
  }

  loading.classList.add('is-hidden');
  window.setTimeout(() => loading.remove(), 220);
}

function showBootError(message: string): void {
  const root = document.querySelector('#game-root');
  if (!root || root.querySelector('.boot-error')) {
    return;
  }

  document.querySelector('#game-loading')?.remove();

  const errorBox = document.createElement('div');
  errorBox.className = 'boot-error';
  errorBox.innerHTML = `
    <div class="boot-error-card">
      <div class="boot-error-title">游戏启动失败</div>
      <div class="boot-error-message">${escapeHtml(message)}</div>
    </div>
  `;
  root.appendChild(errorBox);

  window.FlutterTrainingGame?.postMessage(
    JSON.stringify({
      type: 'training-game-error',
      payload: { message },
    }),
  );
}

function applyGameBootCopy(gameId: string): void {
  if (gameId !== 'firefly-path-trace' && gameId !== 'shadow-theater') {
    return;
  }

  document.title = gameId === 'shadow-theater' ? '影子剧场' : '萤火小路';
  const loadingTitle = document.querySelector('.loading-title');
  if (loadingTitle) {
    loadingTitle.textContent = document.title;
  }
  const loadingText = document.querySelector('.loading-text');
  if (loadingText) {
    loadingText.textContent = '游戏加载中...';
  }
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}
