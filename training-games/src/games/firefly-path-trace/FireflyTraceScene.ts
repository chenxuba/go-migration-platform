import Phaser from 'phaser';
import {
  GameLaunchParams,
  GameResult,
  notifyGameReady,
  requestClose,
  submitGameResult,
} from '../../platform/hostBridge';
import { publicAssetPath } from '../../platform/publicPath';

interface TraceEvent {
  round: number;
  targetColor: string;
  pickedColor: string;
  correct: boolean;
  reactionMs: number;
}

interface LevelPath {
  points: Phaser.Math.Vector2[];
  label: string;
}

const GAME_WIDTH = 1280;
const GAME_HEIGHT = 720;
const ASSET_BASE = publicAssetPath('assets/firefly-trace');
const BACKGROUND_TEXTURE = 'firefly-trace-background';
const GUIDE_TEXTURE = 'firefly-trace-guide';
const ORB_TEXTURE = 'firefly-trace-orb';
const STAR_TEXTURE = 'firefly-trace-star';

const PATHS: LevelPath[] = [
  {
    label: '弯弯小路',
    points: [
      new Phaser.Math.Vector2(270, 470),
      new Phaser.Math.Vector2(410, 360),
      new Phaser.Math.Vector2(580, 430),
      new Phaser.Math.Vector2(740, 300),
      new Phaser.Math.Vector2(930, 378),
      new Phaser.Math.Vector2(1040, 265),
    ],
  },
  {
    label: '星光波浪',
    points: [
      new Phaser.Math.Vector2(250, 350),
      new Phaser.Math.Vector2(390, 262),
      new Phaser.Math.Vector2(530, 396),
      new Phaser.Math.Vector2(690, 270),
      new Phaser.Math.Vector2(850, 410),
      new Phaser.Math.Vector2(1030, 318),
    ],
  },
  {
    label: '安静山丘',
    points: [
      new Phaser.Math.Vector2(270, 520),
      new Phaser.Math.Vector2(390, 430),
      new Phaser.Math.Vector2(520, 352),
      new Phaser.Math.Vector2(675, 352),
      new Phaser.Math.Vector2(835, 430),
      new Phaser.Math.Vector2(1015, 338),
    ],
  },
];

export class FireflyTraceScene extends Phaser.Scene {
  private readonly launchParams: GameLaunchParams;
  private readonly roundTotal: number;

  private startedAt = 0;
  private roundStartedAt = 0;
  private roundIndex = 0;
  private correct = 0;
  private wrong = 0;
  private bestCombo = 0;
  private combo = 0;
  private roundEvents: TraceEvent[] = [];
  private lastResult?: GameResult;
  private audioContext?: AudioContext;

  private hud!: Phaser.GameObjects.Container;
  private playLayer!: Phaser.GameObjects.Container;
  private overlayLayer!: Phaser.GameObjects.Container;
  private roundText!: Phaser.GameObjects.Text;
  private scoreText!: Phaser.GameObjects.Text;
  private promptText!: Phaser.GameObjects.Text;
  private orb!: Phaser.GameObjects.Container;
  private guide!: Phaser.GameObjects.Container;
  private pathGraphics!: Phaser.GameObjects.Graphics;
  private progressDots: Phaser.GameObjects.Arc[] = [];
  private checkpoints: Phaser.GameObjects.Container[] = [];
  private currentPath: LevelPath = PATHS[0];
  private checkpointIndex = 0;
  private dragging = false;

  constructor(launchParams: GameLaunchParams) {
    super('FireflyTraceScene');
    this.launchParams = launchParams;
    this.roundTotal = launchParams.difficulty === 'easy' ? 5 : launchParams.difficulty === 'hard' ? 8 : 6;
  }

  preload(): void {
    this.load.image(BACKGROUND_TEXTURE, `${ASSET_BASE}/background/quiet-glow-garden.png`);
    this.load.image(GUIDE_TEXTURE, `${ASSET_BASE}/characters/firefly-guide.png`);
    this.load.image(ORB_TEXTURE, `${ASSET_BASE}/props/glow-orb.png`);
    this.load.image(STAR_TEXTURE, `${ASSET_BASE}/props/glow-star.png`);
  }

  create(): void {
    this.createFallbackTextures();
    this.createWorld();
    this.createHud();
    this.createPlayLayer();
    this.createStartOverlay();

    this.time.delayedCall(80, () => {
      window.__COLOR_MATCH_READY__ = true;
      window.dispatchEvent(new Event('color-match-ready'));
      notifyGameReady();
    });
  }

  update(): void {
    if (!this.dragging) {
      return;
    }
    this.checkCurrentCheckpoint();
  }

  private createWorld(): void {
    if (this.textures.exists(BACKGROUND_TEXTURE)) {
      const background = this.add.image(GAME_WIDTH / 2, GAME_HEIGHT / 2, BACKGROUND_TEXTURE);
      background.setScale(Math.max(GAME_WIDTH / background.width, GAME_HEIGHT / background.height));
      const veil = this.add.graphics();
      veil.fillGradientStyle(0x071d2a, 0x071d2a, 0xffffff, 0xffffff, 0.1, 0.08, 0.02, 0.02);
      veil.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
      return;
    }

    this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x163d4a);
    this.add.rectangle(GAME_WIDTH / 2, 610, GAME_WIDTH, 220, 0x1b5a55);
  }

  private createHud(): void {
    this.hud = this.add.container(0, 0);
    const panel = this.add.graphics();
    panel.fillStyle(0xeffcff, 0.9);
    panel.fillRoundedRect(72, 26, 1136, 76, 24);
    panel.lineStyle(3, 0x8bdde4, 0.8);
    panel.strokeRoundedRect(72, 26, 1136, 76, 24);
    this.hud.add(panel);

    this.roundText = this.add.text(148, 64, `1/${this.roundTotal}`, {
      color: '#24546b',
      fontSize: '24px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 2, right: 2 },
    });
    this.roundText.setOrigin(0.5);
    this.hud.add(this.roundText);

    this.progressDots = [];
    for (let i = 0; i < this.roundTotal; i += 1) {
      const dot = this.add.circle(278 + i * 42, 64, 8, 0x8bdde4, 0.36);
      this.progressDots.push(dot);
      this.hud.add(dot);
    }

    this.scoreText = this.add.text(920, 64, '收集 0', {
      color: '#ffe08a',
      fontSize: '25px',
      fontStyle: '900',
      stroke: '#24546b',
      strokeThickness: 4,
      padding: { top: 8, bottom: 6, left: 2, right: 2 },
    });
    this.scoreText.setOrigin(0.5);
    this.hud.add(this.scoreText);
  }

  private createPlayLayer(): void {
    this.playLayer = this.add.container(0, 0);
    this.overlayLayer = this.add.container(0, 0);
    this.pathGraphics = this.add.graphics();
    this.playLayer.add(this.pathGraphics);

    this.promptText = this.add.text(640, 146, '按住小光点，沿着小路慢慢走', {
      align: 'center',
      color: '#effcff',
      fontSize: '30px',
      fontStyle: '900',
      stroke: '#123746',
      strokeThickness: 5,
      padding: { top: 10, bottom: 6, left: 8, right: 8 },
    });
    this.promptText.setOrigin(0.5);
    this.playLayer.add(this.promptText);

    this.guide = this.createGuide(146, 536);
    this.playLayer.add(this.guide);

    this.orb = this.createOrb(270, 470);
    this.playLayer.add(this.orb);

    this.playLayer.setVisible(false);
    this.hud.setVisible(false);
  }

  private createStartOverlay(): void {
    this.overlayLayer.removeAll(true);
    const shade = this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x071d2a, 0.34);
    const panel = this.add.graphics();
    panel.fillStyle(0xf5fffe, 0.96);
    panel.fillRoundedRect(318, 138, 644, 410, 34);
    panel.lineStyle(5, 0x8bdde4, 1);
    panel.strokeRoundedRect(318, 138, 644, 410, 34);
    const title = this.add.text(640, 218, '萤火小路', {
      align: 'center',
      color: '#24546b',
      fontSize: '56px',
      fontStyle: '900',
      padding: { top: 12, bottom: 8, left: 8, right: 8 },
    });
    title.setOrigin(0.5);
    const subtitle = this.add.text(640, 310, '按住发光小点，沿着小路收集星光', {
      align: 'center',
      color: '#527482',
      fontSize: '28px',
      fontStyle: '800',
      padding: { top: 10, bottom: 6, left: 8, right: 8 },
    });
    subtitle.setOrigin(0.5);
    const start = this.createButton(640, 446, 250, 72, '开始', 0x55c98f, 0x32896b, () => {
      this.ensureAudio();
      this.playTone(660, 0.1, 'sine');
      this.startGame();
    });
    this.overlayLayer.add([shade, panel, title, subtitle, start]);
  }

  private startGame(): void {
    this.overlayLayer.removeAll(true);
    this.playLayer.setVisible(true);
    this.hud.setVisible(true);
    this.startedAt = performance.now();
    this.roundIndex = 0;
    this.correct = 0;
    this.wrong = 0;
    this.combo = 0;
    this.bestCombo = 0;
    this.roundEvents = [];
    this.nextRound();
  }

  private nextRound(): void {
    if (this.roundIndex >= this.roundTotal) {
      this.finishGame();
      return;
    }

    this.currentPath = PATHS[this.roundIndex % PATHS.length];
    this.roundStartedAt = performance.now();
    this.checkpointIndex = 0;
    this.dragging = false;
    this.promptText.setText('按住小光点，沿着小路慢慢走');
    this.drawPath();
    this.createCheckpoints();
    const start = this.currentPath.points[0];
    this.orb.setPosition(start.x, start.y);
    this.updateHud();
  }

  private drawPath(): void {
    this.pathGraphics.clear();
    const points = this.currentPath.points;
    this.pathGraphics.lineStyle(24, 0x123746, 0.16);
    this.pathGraphics.beginPath();
    this.pathGraphics.moveTo(points[0].x, points[0].y);
    for (const point of points.slice(1)) {
      this.pathGraphics.lineTo(point.x, point.y);
    }
    this.pathGraphics.strokePath();
    this.pathGraphics.lineStyle(10, 0xffe08a, 0.42);
    this.pathGraphics.beginPath();
    this.pathGraphics.moveTo(points[0].x, points[0].y);
    for (const point of points.slice(1)) {
      this.pathGraphics.lineTo(point.x, point.y);
    }
    this.pathGraphics.strokePath();
  }

  private createCheckpoints(): void {
    for (const checkpoint of this.checkpoints) {
      checkpoint.destroy();
    }
    this.checkpoints = [];
    this.currentPath.points.slice(1).forEach((point, index) => {
      const checkpoint = this.add.container(point.x, point.y);
      const glow = this.add.circle(0, 0, 30, 0xffe08a, 0.18);
      const star = this.add.image(0, 0, this.textures.exists(STAR_TEXTURE) ? STAR_TEXTURE : 'trace-star');
      star.setDisplaySize(44, 44);
      checkpoint.add([glow, star]);
      checkpoint.setAlpha(index === 0 ? 1 : 0.42);
      this.playLayer.add(checkpoint);
      this.checkpoints.push(checkpoint);
    });
  }

  private createOrb(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const glow = this.add.circle(0, 0, 48, 0xffe08a, 0.18);
    const orb = this.add.image(0, 0, this.textures.exists(ORB_TEXTURE) ? ORB_TEXTURE : 'trace-orb');
    orb.setDisplaySize(86, 86);
    const hit = this.add.zone(0, 0, 104, 104);
    hit.setOrigin(0.5);
    hit.setInteractive({ draggable: true, useHandCursor: true });
    hit.on('dragstart', () => {
      this.dragging = true;
      this.promptText.setText('很好，慢慢跟着光走');
    });
    hit.on('drag', (pointer: Phaser.Input.Pointer) => {
      container.setPosition(
        Phaser.Math.Clamp(pointer.worldX, 120, 1160),
        Phaser.Math.Clamp(pointer.worldY, 150, 620),
      );
    });
    hit.on('dragend', () => {
      this.dragging = false;
      if (this.checkpointIndex < this.checkpoints.length) {
        this.promptText.setText('继续按住小光点，回到小路上');
      }
    });
    container.add([glow, orb, hit]);
    this.tweens.add({ targets: glow, scale: 1.16, alpha: 0.32, duration: 820, yoyo: true, repeat: -1 });
    return container;
  }

  private checkCurrentCheckpoint(): void {
    const checkpoint = this.checkpoints[this.checkpointIndex];
    if (!checkpoint) {
      return;
    }
    const distance = Phaser.Math.Distance.Between(this.orb.x, this.orb.y, checkpoint.x, checkpoint.y);
    if (distance > 46) {
      return;
    }

    checkpoint.setAlpha(1);
    this.tweens.add({ targets: checkpoint, scale: 1.24, alpha: 0, duration: 260, ease: 'Back.Out' });
    this.playTone(700 + this.checkpointIndex * 60, 0.08, 'sine');
    this.checkpointIndex += 1;

    const next = this.checkpoints[this.checkpointIndex];
    if (next) {
      next.setAlpha(1);
      return;
    }

    this.completeRound();
  }

  private completeRound(): void {
    this.dragging = false;
    this.correct += 1;
    this.combo += 1;
    this.bestCombo = Math.max(this.bestCombo, this.combo);
    this.roundEvents.push({
      round: this.roundIndex + 1,
      targetColor: this.currentPath.label,
      pickedColor: 'completed',
      correct: true,
      reactionMs: Math.round(performance.now() - this.roundStartedAt),
    });
    this.promptText.setText('收集完成');
    this.tweens.add({ targets: this.orb, scale: 1.18, duration: 160, yoyo: true, ease: 'Back.Out' });
    this.updateHud();
    this.time.delayedCall(720, () => {
      this.roundIndex += 1;
      this.nextRound();
    });
  }

  private updateHud(): void {
    this.roundText.setText(`${Math.min(this.roundIndex + 1, this.roundTotal)}/${this.roundTotal}`);
    this.scoreText.setText(`收集 ${this.correct}`);
    this.progressDots.forEach((dot, index) => {
      dot.setFillStyle(index < this.roundIndex ? 0xffe08a : index === this.roundIndex ? 0x55c98f : 0x8bdde4);
      dot.setAlpha(index <= this.roundIndex ? 1 : 0.36);
      dot.setScale(index === this.roundIndex ? 1.24 : 1);
    });
  }

  private finishGame(): void {
    const durationMs = Math.round(performance.now() - this.startedAt);
    const avgReactionMs =
      this.roundEvents.length === 0 ? 0 : Math.round(this.roundEvents.reduce((sum, event) => sum + event.reactionMs, 0) / this.roundEvents.length);
    const accuracy = this.roundTotal === 0 ? 0 : Number((this.correct / this.roundTotal).toFixed(2));
    const stars = this.correct >= this.roundTotal ? 3 : accuracy >= 0.75 ? 2 : accuracy >= 0.5 ? 1 : 0;
    this.lastResult = {
      gameId: this.launchParams.gameId,
      taskId: this.launchParams.taskId,
      studentId: this.launchParams.studentId,
      startedAt: new Date(Date.now() - durationMs).toISOString(),
      endedAt: new Date().toISOString(),
      durationMs,
      total: this.roundTotal,
      correct: this.correct,
      wrong: this.wrong,
      accuracy,
      bestCombo: this.bestCombo,
      avgReactionMs,
      stars,
      events: this.roundEvents,
    };
    void submitGameResult(this.lastResult, this.launchParams.token);
    this.showResultOverlay(this.lastResult);
  }

  private showResultOverlay(result: GameResult): void {
    this.playLayer.setVisible(false);
    this.hud.setVisible(false);
    this.overlayLayer.removeAll(true);
    const shade = this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x071d2a, 0.42);
    const panel = this.add.graphics();
    panel.fillStyle(0xf5fffe, 0.96);
    panel.fillRoundedRect(286, 104, 708, 512, 34);
    panel.lineStyle(5, 0x8bdde4, 1);
    panel.strokeRoundedRect(286, 104, 708, 512, 34);
    const title = this.add.text(640, 174, result.stars >= 2 ? '小路完成！' : '练习完成！', {
      color: '#24546b',
      fontSize: '50px',
      fontStyle: '900',
      padding: { top: 12, bottom: 8, left: 8, right: 8 },
    });
    title.setOrigin(0.5);
    const stats = this.add.text(640, 330, `完成 ${result.correct} / ${result.total}\n平均用时 ${Math.round(result.avgReactionMs / 100) / 10} 秒`, {
      align: 'center',
      color: '#41596b',
      fontSize: '30px',
      fontStyle: '800',
      lineSpacing: 18,
      padding: { top: 10, bottom: 8, left: 8, right: 8 },
    });
    stats.setOrigin(0.5);
    const replay = this.createButton(514, 510, 210, 68, '再玩一次', 0x55c98f, 0x32896b, () => this.startGame());
    const done = this.createButton(766, 510, 210, 68, '完成', 0xffaa62, 0xd98538, () => requestClose(this.lastResult));
    this.overlayLayer.add([shade, panel, title, stats, replay, done]);
  }

  private createGuide(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const shadow = this.add.ellipse(4, 78, 116, 26, 0x071d2a, 0.18);
    if (this.textures.exists(GUIDE_TEXTURE)) {
      const guide = this.add.image(0, 0, GUIDE_TEXTURE);
      guide.setDisplaySize(176, 176);
      container.add([shadow, guide]);
    } else {
      const guide = this.add.image(0, 0, 'trace-orb');
      guide.setScale(1.8);
      container.add([shadow, guide]);
    }
    this.tweens.add({ targets: container, y: y - 10, duration: 1200, yoyo: true, repeat: -1, ease: 'Sine.InOut' });
    return container;
  }

  private createButton(
    x: number,
    y: number,
    width: number,
    height: number,
    label: string,
    color: number,
    shadowColor: number,
    onClick: () => void,
  ): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const shadow = this.add.graphics();
    shadow.fillStyle(shadowColor, 1);
    shadow.fillRoundedRect(-width / 2, -height / 2 + 8, width, height, 22);
    const body = this.add.graphics();
    body.fillStyle(color, 1);
    body.fillRoundedRect(-width / 2, -height / 2, width, height, 22);
    body.lineStyle(4, 0xffffff, 0.55);
    body.strokeRoundedRect(-width / 2, -height / 2, width, height, 22);
    const text = this.add.text(0, -2, label, {
      align: 'center',
      color: '#ffffff',
      fontSize: '30px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 6, right: 6 },
    });
    text.setOrigin(0.5);
    const hitZone = this.add.zone(0, 4, width, height + 16);
    hitZone.setOrigin(0.5);
    hitZone.setInteractive({ useHandCursor: true });
    hitZone.on('pointerdown', () => this.tweens.add({ targets: container, scale: 0.94, duration: 70, yoyo: true, onComplete: onClick }));
    container.add([shadow, body, text, hitZone]);
    return container;
  }

  private ensureAudio(): void {
    if (!this.audioContext) {
      const AudioContextClass = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
      this.audioContext = new AudioContextClass();
    }
    if (this.audioContext.state === 'suspended') {
      void this.audioContext.resume();
    }
  }

  private playTone(frequency: number, duration: number, type: OscillatorType): void {
    this.ensureAudio();
    if (!this.audioContext) {
      return;
    }
    const start = this.audioContext.currentTime;
    const oscillator = this.audioContext.createOscillator();
    const gain = this.audioContext.createGain();
    oscillator.type = type;
    oscillator.frequency.value = frequency;
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(0.1, start + 0.015);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + duration);
    oscillator.connect(gain);
    gain.connect(this.audioContext.destination);
    oscillator.start(start);
    oscillator.stop(start + duration + 0.03);
  }

  private createFallbackTextures(): void {
    this.makeGlowCircle('trace-orb', 0xffe08a);
    this.makeStar('trace-star', 0xffe08a);
  }

  private makeGlowCircle(key: string, color: number): void {
    const graphics = this.add.graphics();
    graphics.fillStyle(color, 0.16);
    graphics.fillCircle(44, 44, 42);
    graphics.fillStyle(color, 0.88);
    graphics.fillCircle(44, 44, 18);
    graphics.generateTexture(key, 88, 88);
    graphics.destroy();
  }

  private makeStar(key: string, color: number): void {
    const graphics = this.add.graphics();
    const points: Phaser.Math.Vector2[] = [];
    for (let i = 0; i < 10; i += 1) {
      const radius = i % 2 === 0 ? 24 : 10;
      const angle = Phaser.Math.DegToRad(i * 36 - 90);
      points.push(new Phaser.Math.Vector2(30 + Math.cos(angle) * radius, 30 + Math.sin(angle) * radius));
    }
    graphics.fillStyle(color, 1);
    graphics.lineStyle(4, 0xffffff, 0.8);
    graphics.beginPath();
    graphics.moveTo(points[0].x, points[0].y);
    for (const point of points.slice(1)) {
      graphics.lineTo(point.x, point.y);
    }
    graphics.closePath();
    graphics.fillPath();
    graphics.strokePath();
    graphics.generateTexture(key, 60, 60);
    graphics.destroy();
  }
}
