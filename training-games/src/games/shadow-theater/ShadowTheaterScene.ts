import Phaser from 'phaser';
import {
  GameLaunchParams,
  GameResult,
  notifyGameReady,
  requestClose,
  submitGameResult,
} from '../../platform/hostBridge';

interface PuppetOption {
  key: string;
  name: string;
  tint: number;
}

interface ShadowRound {
  puppet: PuppetOption;
  targetAngle: number;
  targetScale: number;
}

interface ShadowEvent {
  round: number;
  targetColor: string;
  pickedColor: string;
  correct: boolean;
  reactionMs: number;
}

const GAME_WIDTH = 1280;
const GAME_HEIGHT = 720;
const SHADOW_CENTER_X = 640;
const SHADOW_CENTER_Y = 338;
const ASSET_BASE = '/assets/shadow-theater';
const BACKGROUND_TEXTURE = 'shadow-theater-background';
const PUPPET_PREFIX = 'shadow-theater-puppet';

const PUPPETS: PuppetOption[] = [
  { key: 'butterfly', name: '蝴蝶', tint: 0xff9c84 },
  { key: 'kite', name: '风筝', tint: 0x81d4f5 },
  { key: 'boat', name: '小船', tint: 0xffc66e },
  { key: 'moon', name: '月亮', tint: 0xffdc6f },
];

const ANGLES = [-42, -24, 18, 36, 58, -66];
const SCALES = [0.76, 0.88, 1, 1.12, 1.24];

export class ShadowTheaterScene extends Phaser.Scene {
  private readonly launchParams: GameLaunchParams;
  private readonly roundTotal: number;

  private startedAt = 0;
  private roundStartedAt = 0;
  private roundIndex = 0;
  private correct = 0;
  private wrong = 0;
  private combo = 0;
  private bestCombo = 0;
  private soundEnabled = true;
  private roundEvents: ShadowEvent[] = [];
  private lastResult?: GameResult;
  private audioContext?: AudioContext;

  private hud!: Phaser.GameObjects.Container;
  private playLayer!: Phaser.GameObjects.Container;
  private overlayLayer!: Phaser.GameObjects.Container;
  private roundText!: Phaser.GameObjects.Text;
  private scoreText!: Phaser.GameObjects.Text;
  private promptText!: Phaser.GameObjects.Text;
  private hintText!: Phaser.GameObjects.Text;
  private soundText!: Phaser.GameObjects.Text;
  private shadowImage!: Phaser.GameObjects.Image;
  private puppetImage!: Phaser.GameObjects.Image;
  private checkButton!: Phaser.GameObjects.Container;
  private progressDots: Phaser.GameObjects.Arc[] = [];
  private currentRound!: ShadowRound;
  private puppetAngle = 0;
  private puppetScale = 1;
  private dragStartX = 0;
  private dragStartY = 0;
  private dragStartAngle = 0;
  private dragStartScale = 1;

  constructor(launchParams: GameLaunchParams) {
    super('ShadowTheaterScene');
    this.launchParams = launchParams;
    this.roundTotal = launchParams.difficulty === 'easy' ? 6 : launchParams.difficulty === 'hard' ? 10 : 8;
  }

  preload(): void {
    this.load.image(BACKGROUND_TEXTURE, `${ASSET_BASE}/background/shadow-stage.png`);
    for (const puppet of PUPPETS) {
      this.load.image(this.puppetKey(puppet), `${ASSET_BASE}/puppets/${puppet.key}.png`);
    }
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

  private createWorld(): void {
    if (this.textures.exists(BACKGROUND_TEXTURE)) {
      const bg = this.add.image(GAME_WIDTH / 2, GAME_HEIGHT / 2, BACKGROUND_TEXTURE);
      bg.setScale(Math.max(GAME_WIDTH / bg.width, GAME_HEIGHT / bg.height));
      this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x120d1b, 0.06);
      return;
    }
    this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x1d2638);
    this.add.rectangle(640, 264, 720, 300, 0xfff3c8, 0.9);
  }

  private createHud(): void {
    this.hud = this.add.container(0, 0);
    const panel = this.add.graphics();
    panel.fillStyle(0xfffbeb, 0.92);
    panel.fillRoundedRect(76, 24, 1128, 76, 24);
    panel.lineStyle(3, 0xf5c66b, 0.9);
    panel.strokeRoundedRect(76, 24, 1128, 76, 24);
    this.hud.add(panel);

    this.roundText = this.add.text(148, 64, `1/${this.roundTotal}`, {
      color: '#47301c',
      fontSize: '24px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 2, right: 2 },
    });
    this.roundText.setOrigin(0.5);
    this.hud.add(this.roundText);

    this.progressDots = [];
    for (let i = 0; i < this.roundTotal; i += 1) {
      const dot = this.add.circle(270 + i * 38, 64, 7.5, 0xf3d6a1, 1);
      this.progressDots.push(dot);
      this.hud.add(dot);
    }

    this.scoreText = this.add.text(856, 64, '完成 0', {
      color: '#7c4b00',
      fontSize: '24px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 2, right: 2 },
    });
    this.scoreText.setOrigin(0.5);
    this.hud.add(this.scoreText);

    this.hud.add(this.createSoundToggle(1082, 64));
  }

  private createPlayLayer(): void {
    this.playLayer = this.add.container(0, 0);
    this.overlayLayer = this.add.container(0, 0);

    const screenGlow = this.add.graphics();
    screenGlow.fillStyle(0xffe7a6, 0.13);
    screenGlow.fillRoundedRect(332, 152, 616, 376, 26);
    screenGlow.lineStyle(2, 0xfff2bc, 0.22);
    screenGlow.strokeRoundedRect(332, 152, 616, 376, 26);
    this.playLayer.add(screenGlow);

    this.promptText = this.add.text(640, 126, '拖动纸偶，让它盖住影子', {
      align: 'center',
      color: '#fff5cc',
      fontSize: '28px',
      fontStyle: '900',
      stroke: '#47301c',
      strokeThickness: 4,
      padding: { top: 10, bottom: 6, left: 8, right: 8 },
    });
    this.promptText.setOrigin(0.5);
    this.playLayer.add(this.promptText);

    this.shadowImage = this.add.image(SHADOW_CENTER_X, SHADOW_CENTER_Y, 'shadow-fallback');
    this.shadowImage.setAlpha(0.42);
    this.shadowImage.setTint(0x1b1724);
    this.playLayer.add(this.shadowImage);

    this.puppetImage = this.add.image(SHADOW_CENTER_X, SHADOW_CENTER_Y, 'shadow-fallback');
    this.puppetImage.setInteractive({ draggable: true, useHandCursor: true });
    this.puppetImage.on('dragstart', (pointer: Phaser.Input.Pointer) => {
      this.dragStartX = pointer.worldX;
      this.dragStartY = pointer.worldY;
      this.dragStartAngle = this.puppetAngle;
      this.dragStartScale = this.puppetScale;
      this.hintText.setText('左右转动，上下调整大小');
    });
    this.puppetImage.on('drag', (pointer: Phaser.Input.Pointer) => {
      const dx = pointer.worldX - this.dragStartX;
      const dy = pointer.worldY - this.dragStartY;
      this.puppetAngle = Phaser.Math.Clamp(this.dragStartAngle + dx * 0.32, -85, 85);
      this.puppetScale = Phaser.Math.Clamp(this.dragStartScale - dy * 0.004, 0.66, 1.34);
      this.applyPuppetTransform();
      this.updateLiveHint();
    });
    this.playLayer.add(this.puppetImage);

    this.hintText = this.add.text(576, 626, '左右拖动旋转，上下拖动变大变小', {
      align: 'center',
      color: '#fff5cc',
      fontSize: '24px',
      fontStyle: '800',
      stroke: '#47301c',
      strokeThickness: 4,
      padding: { top: 8, bottom: 6, left: 8, right: 8 },
    });
    this.hintText.setOrigin(0.5);
    this.playLayer.add(this.hintText);

    this.checkButton = this.createButton(1052, 628, 178, 60, '完成', 0xf5a84f, 0xc8782e, () => this.checkAnswer());
    this.playLayer.add(this.checkButton);

    this.playLayer.setVisible(false);
    this.hud.setVisible(false);
  }

  private createStartOverlay(): void {
    this.overlayLayer.removeAll(true);
    const shade = this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x120d1b, 0.36);
    const panel = this.add.graphics();
    panel.fillStyle(0xfffbeb, 0.96);
    panel.fillRoundedRect(318, 140, 644, 398, 34);
    panel.lineStyle(5, 0xf5c66b, 1);
    panel.strokeRoundedRect(318, 140, 644, 398, 34);
    const title = this.add.text(640, 220, '影子剧场', {
      color: '#47301c',
      fontSize: '56px',
      fontStyle: '900',
      padding: { top: 12, bottom: 8, left: 8, right: 8 },
    });
    title.setOrigin(0.5);
    const text = this.add.text(640, 316, '拖动纸偶，调整方向和大小\\n让它和幕布影子重合', {
      align: 'center',
      color: '#6c5238',
      fontSize: '28px',
      fontStyle: '800',
      lineSpacing: 10,
      padding: { top: 10, bottom: 8, left: 8, right: 8 },
    });
    text.setOrigin(0.5);
    const start = this.createButton(640, 456, 250, 72, '开始', 0xf5a84f, 0xc8782e, () => {
      if (this.soundEnabled) {
        this.ensureAudio();
      }
      this.playTone(660, 0.1, 'sine');
      this.startGame();
    });
    this.overlayLayer.add([shade, panel, title, text, start]);
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
    const puppet = PUPPETS[this.roundIndex % PUPPETS.length];
    this.currentRound = {
      puppet,
      targetAngle: ANGLES[this.roundIndex % ANGLES.length],
      targetScale: SCALES[(this.roundIndex + 1) % SCALES.length],
    };
    this.roundStartedAt = performance.now();
    const texture = this.textures.exists(this.puppetKey(puppet)) ? this.puppetKey(puppet) : 'shadow-fallback';
    this.shadowImage.setTexture(texture);
    this.fitImageToBox(this.shadowImage, 270, 270, this.currentRound.targetScale);
    this.shadowImage.setPosition(SHADOW_CENTER_X, SHADOW_CENTER_Y);
    this.shadowImage.setAngle(this.currentRound.targetAngle);
    this.shadowImage.setTint(0x171321);
    this.shadowImage.setAlpha(0.42);
    this.puppetImage.setTexture(texture);
    this.puppetImage.setPosition(SHADOW_CENTER_X, SHADOW_CENTER_Y);
    this.puppetImage.setTint(0xffffff);
    this.puppetImage.setAlpha(0.94);
    this.puppetAngle = Phaser.Math.Clamp(this.currentRound.targetAngle + Phaser.Utils.Array.GetRandom([-46, -32, 34, 48]), -85, 85);
    this.puppetScale = Phaser.Math.Clamp(this.currentRound.targetScale + Phaser.Utils.Array.GetRandom([-0.22, -0.16, 0.18, 0.24]), 0.66, 1.34);
    this.applyPuppetTransform();
    this.promptText.setText(`拖动${puppet.name}，盖住幕布影子`);
    this.hintText.setText('左右拖动旋转，上下拖动变大变小');
    this.updateHud();
  }

  private applyPuppetTransform(): void {
    this.puppetImage.setAngle(this.puppetAngle);
    this.fitImageToBox(this.puppetImage, 270, 270, this.puppetScale);
  }

  private updateLiveHint(): void {
    const score = this.matchScore();
    if (score >= 0.86) {
      this.hintText.setText('很接近了，点完成试试');
    } else if (score >= 0.68) {
      this.hintText.setText('快到了，再微调一点');
    } else {
      this.hintText.setText('左右调方向，上下调大小');
    }
  }

  private checkAnswer(): void {
    const score = this.matchScore();
    const correct = score >= 0.8;
    const reactionMs = Math.round(performance.now() - this.roundStartedAt);
    this.roundEvents.push({
      round: this.roundIndex + 1,
      targetColor: this.currentRound.puppet.key,
      pickedColor: `${Math.round(this.puppetAngle)}:${this.puppetScale.toFixed(2)}`,
      correct,
      reactionMs,
    });

    if (correct) {
      this.correct += 1;
      this.combo += 1;
      this.bestCombo = Math.max(this.bestCombo, this.combo);
      this.playTone(784, 0.08, 'sine');
      this.playTone(1046, 0.12, 'triangle');
      this.tweens.add({ targets: this.puppetImage, alpha: 1, duration: 120, yoyo: true, repeat: 1 });
      this.tweens.add({ targets: this.shadowImage, alpha: 0.12, duration: 260 });
      this.hintText.setText('重合成功');
      this.updateHud();
      this.time.delayedCall(760, () => {
        this.roundIndex += 1;
        this.nextRound();
      });
      return;
    }

    this.wrong += 1;
    this.combo = 0;
    this.playTone(260, 0.12, 'sine');
    this.hintText.setText('还差一点，慢慢调');
    this.tweens.add({ targets: this.shadowImage, alpha: 0.66, duration: 150, yoyo: true });
    this.updateHud();
  }

  private matchScore(): number {
    const angleDiff = Math.abs(Phaser.Math.Angle.ShortestBetween(this.puppetAngle, this.currentRound.targetAngle));
    const scaleDiff = Math.abs(this.puppetScale - this.currentRound.targetScale);
    const angleScore = Math.max(0, 1 - angleDiff / 30);
    const scaleScore = Math.max(0, 1 - scaleDiff / 0.26);
    return angleScore * 0.58 + scaleScore * 0.42;
  }

  private updateHud(): void {
    this.roundText.setText(`${Math.min(this.roundIndex + 1, this.roundTotal)}/${this.roundTotal}`);
    this.scoreText.setText(`完成 ${this.correct}`);
    this.progressDots.forEach((dot, index) => {
      dot.setFillStyle(index < this.roundIndex ? 0xf5a84f : index === this.roundIndex ? 0xffdc6f : 0xf3d6a1);
      dot.setScale(index === this.roundIndex ? 1.25 : 1);
    });
  }

  private finishGame(): void {
    const durationMs = Math.round(performance.now() - this.startedAt);
    const avgReactionMs =
      this.roundEvents.length === 0
        ? 0
        : Math.round(this.roundEvents.reduce((sum, event) => sum + event.reactionMs, 0) / this.roundEvents.length);
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
    const shade = this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x120d1b, 0.46);
    const panel = this.add.graphics();
    panel.fillStyle(0xfffbeb, 0.97);
    panel.fillRoundedRect(286, 104, 708, 512, 34);
    panel.lineStyle(5, 0xf5c66b, 1);
    panel.strokeRoundedRect(286, 104, 708, 512, 34);
    const title = this.add.text(640, 178, result.stars >= 2 ? '剧场完成！' : '练习完成！', {
      color: '#47301c',
      fontSize: '50px',
      fontStyle: '900',
      padding: { top: 12, bottom: 8, left: 8, right: 8 },
    });
    title.setOrigin(0.5);
    const stats = this.add.text(640, 328, `完成 ${result.correct} / ${result.total}\n平均用时 ${Math.round(result.avgReactionMs / 100) / 10} 秒`, {
      align: 'center',
      color: '#6c5238',
      fontSize: '30px',
      fontStyle: '800',
      lineSpacing: 18,
      padding: { top: 10, bottom: 8, left: 8, right: 8 },
    });
    stats.setOrigin(0.5);
    const replay = this.createButton(514, 510, 210, 68, '再玩一次', 0xf5a84f, 0xc8782e, () => this.startGame());
    const done = this.createButton(766, 510, 210, 68, '完成', 0x55c98f, 0x32896b, () => requestClose(this.lastResult));
    this.overlayLayer.add([shade, panel, title, stats, replay, done]);
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
      color: '#ffffff',
      fontSize: '30px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 6, right: 6 },
    });
    text.setOrigin(0.5);
    const hit = this.add.zone(0, 4, width, height + 16);
    hit.setOrigin(0.5);
    hit.setInteractive({ useHandCursor: true });
    hit.on('pointerdown', () => this.tweens.add({ targets: container, scale: 0.94, duration: 70, yoyo: true, onComplete: onClick }));
    container.add([shadow, body, text, hit]);
    return container;
  }

  private createSoundToggle(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const shadow = this.add.graphics();
    shadow.fillStyle(0xc8914a, 0.9);
    shadow.fillRoundedRect(-76, -22 + 5, 152, 44, 18);
    const body = this.add.graphics();
    body.fillStyle(0xfff6d8, 0.98);
    body.fillRoundedRect(-76, -22, 152, 44, 18);
    body.lineStyle(3, 0xf2c06a, 0.9);
    body.strokeRoundedRect(-76, -22, 152, 44, 18);
    const icon = this.add.graphics();
    icon.fillStyle(0x7c4b00, 1);
    icon.fillRoundedRect(-58, -8, 9, 16, 2);
    icon.fillTriangle(-49, -12, -34, -2, -49, 8);
    icon.lineStyle(3, 0x7c4b00, 1);
    icon.beginPath();
    icon.arc(-30, -1, 9, -0.78, 0.78);
    icon.strokePath();
    this.soundText = this.add.text(22, 1, '音效 开', {
      color: '#7c4b00',
      fontSize: '20px',
      fontStyle: '900',
      padding: { top: 6, bottom: 4, left: 2, right: 2 },
    });
    this.soundText.setOrigin(0.5);
    const hit = this.add.zone(0, 2, 160, 54);
    hit.setOrigin(0.5);
    hit.setInteractive({ useHandCursor: true });
    hit.on('pointerdown', () => {
      this.soundEnabled = !this.soundEnabled;
      this.updateSoundText();
      if (this.soundEnabled) {
        this.playTone(620, 0.08, 'sine');
      }
    });
    container.add([shadow, body, icon, this.soundText, hit]);
    return container;
  }

  private updateSoundText(): void {
    this.soundText.setText(this.soundEnabled ? '音效 开' : '音效 关');
  }

  private fitImageToBox(
    image: Phaser.GameObjects.Image,
    maxWidth: number,
    maxHeight: number,
    multiplier: number,
  ): void {
    const scale = Math.min(maxWidth / image.width, maxHeight / image.height) * multiplier;
    image.setScale(scale);
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
    if (!this.soundEnabled) {
      return;
    }
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
    const graphics = this.add.graphics();
    graphics.fillStyle(0xf5a84f, 1);
    graphics.fillTriangle(20, 90, 90, 20, 160, 90);
    graphics.fillRoundedRect(42, 86, 96, 48, 12);
    graphics.generateTexture('shadow-fallback', 180, 160);
    graphics.destroy();
  }

  private puppetKey(puppet: PuppetOption): string {
    return `${PUPPET_PREFIX}-${puppet.key}`;
  }
}
