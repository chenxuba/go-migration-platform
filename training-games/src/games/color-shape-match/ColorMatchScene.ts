import Phaser from 'phaser';
import {
  GameLaunchParams,
  GameResult,
  notifyGameReady,
  requestClose,
  submitGameResult,
} from '../../platform/hostBridge';
import { publicAssetPath } from '../../platform/publicPath';

interface ColorOption {
  key: string;
  name: string;
  value: number;
  shadow: number;
}

interface RoundEvent {
  round: number;
  targetColor: string;
  pickedColor: string;
  correct: boolean;
  reactionMs: number;
}

interface BubbleChoice {
  color: ColorOption;
  container: Phaser.GameObjects.Container;
  baseY: number;
}

const COLORS: ColorOption[] = [
  { key: 'red', name: '红色', value: 0xff5a65, shadow: 0xd84350 },
  { key: 'yellow', name: '黄色', value: 0xffd447, shadow: 0xd99d13 },
  { key: 'blue', name: '蓝色', value: 0x4e9cff, shadow: 0x2767c7 },
  { key: 'green', name: '绿色', value: 0x54ce73, shadow: 0x279a4a },
  { key: 'purple', name: '紫色', value: 0xa678ff, shadow: 0x704ccf },
  { key: 'orange', name: '橙色', value: 0xff9b42, shadow: 0xd56517 },
];

const GAME_WIDTH = 1280;
const GAME_HEIGHT = 720;
const ASSET_BASE = publicAssetPath('assets/color-match');
const BACKGROUND_TEXTURE = 'color-match-background';
const MASCOT_TEXTURE = 'color-match-mascot';
const TOKEN_TEXTURE_PREFIX = 'color-match-token';
const CHOICE_CUSHION_TEXTURE = 'color-match-choice-cushion';
const HUD_BAR_TEXTURE = 'color-match-ui-hud-bar';
const TASK_CARD_TEXTURE = 'color-match-ui-task-card';
const LISTEN_BUTTON_TEXTURE = 'color-match-ui-listen-button';
const MUSIC_TOGGLE_ON_TEXTURE = 'color-match-ui-music-toggle-on';
const MUSIC_TOGGLE_OFF_TEXTURE = 'color-match-ui-music-toggle-off';

export class ColorMatchScene extends Phaser.Scene {
  private readonly launchParams: GameLaunchParams;
  private readonly roundTotal: number;
  private readonly choiceTotal: number;

  private startedAt = 0;
  private roundStartedAt = 0;
  private roundIndex = 0;
  private correct = 0;
  private wrong = 0;
  private combo = 0;
  private bestCombo = 0;
  private currentTarget: ColorOption = COLORS[0];
  private lastResult?: GameResult;
  private roundEvents: RoundEvent[] = [];
  private choices: BubbleChoice[] = [];
  private locked = false;
  private audioContext?: AudioContext;
  private musicGain?: GainNode;
  private musicTimer?: number;
  private musicEnabled = true;
  private activeVoice?: HTMLAudioElement;
  private voiceClips = new Map<string, HTMLAudioElement>();
  private readonly stopAudioHandler = (): void => this.stopAllAudio();

  private hud!: Phaser.GameObjects.Container;
  private playLayer!: Phaser.GameObjects.Container;
  private overlayLayer!: Phaser.GameObjects.Container;
  private roundText!: Phaser.GameObjects.Text;
  private scoreText!: Phaser.GameObjects.Text;
  private comboText!: Phaser.GameObjects.Text;
  private musicButton!: Phaser.GameObjects.Container;
  private feedbackText!: Phaser.GameObjects.Text;
  private mascot!: Phaser.GameObjects.Container;
  private targetBubble!: Phaser.GameObjects.Container;
  private taskText!: Phaser.GameObjects.Text;
  private voiceButton!: Phaser.GameObjects.Container;
  private currentPrompt = '';
  private progressDots: Phaser.GameObjects.Arc[] = [];

  constructor(launchParams: GameLaunchParams) {
    super('ColorMatchScene');
    this.launchParams = launchParams;
    this.roundTotal = launchParams.difficulty === 'easy' ? 8 : launchParams.difficulty === 'hard' ? 12 : 10;
    this.choiceTotal = launchParams.difficulty === 'easy' ? 4 : launchParams.difficulty === 'hard' ? 6 : 5;
  }

  preload(): void {
    this.load.image(BACKGROUND_TEXTURE, `${ASSET_BASE}/background/color-playroom.png`);
    this.load.image(MASCOT_TEXTURE, `${ASSET_BASE}/characters/paint-star-guide.png`);
    this.load.image(CHOICE_CUSHION_TEXTURE, `${ASSET_BASE}/props/choice-cushion.png`);
    this.load.image(HUD_BAR_TEXTURE, `${ASSET_BASE}/ui/hud-bar.png`);
    this.load.image(TASK_CARD_TEXTURE, `${ASSET_BASE}/ui/task-card.png`);
    this.load.image(LISTEN_BUTTON_TEXTURE, `${ASSET_BASE}/ui/listen-button.png`);
    this.load.image(MUSIC_TOGGLE_ON_TEXTURE, `${ASSET_BASE}/ui/music-toggle-on.png`);
    this.load.image(MUSIC_TOGGLE_OFF_TEXTURE, `${ASSET_BASE}/ui/music-toggle-off.png`);
    for (const color of COLORS) {
      this.load.image(this.getTokenTextureKey(color), `${ASSET_BASE}/tokens/${color.key}.png`);
    }
  }

  create(): void {
    window.__COLOR_MATCH_STOP_AUDIO__ = this.stopAudioHandler;
    window.addEventListener('pagehide', this.stopAudioHandler);
    window.addEventListener('beforeunload', this.stopAudioHandler);
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        this.stopAllAudio();
      }
    });
    this.musicEnabled = window.localStorage.getItem('colorMatchMusic') !== 'off';
    this.createGeneratedTextures();
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

  update(_time: number, delta: number): void {
    for (const choice of this.choices) {
      choice.container.y += Math.sin((this.time.now + choice.container.x * 5) / 520) * delta * 0.006;
      choice.container.y = Phaser.Math.Linear(choice.container.y, choice.baseY, 0.018);
    }
  }

  private createGeneratedTextures(): void {
    this.makeCircleTexture('spark', 10, 0xffffff);
    this.makeStarTexture('star', 0xffd447, 0xfff0a8);
    this.makeStarTexture('star-white', 0xffffff, 0xcff7ff);
    this.makeBoltTexture('bolt-icon', 0xff6b8a, 0xffffff);
    this.makeMusicTexture('music-icon', 0x4e9cff, 0xffffff);
    this.makeMusicTexture('music-muted-icon', 0xa7b8c2, 0xffffff);
  }

  private createWorld(): void {
    if (this.textures.exists(BACKGROUND_TEXTURE)) {
      const background = this.add.image(GAME_WIDTH / 2, GAME_HEIGHT / 2, BACKGROUND_TEXTURE);
      const scale = Math.max(GAME_WIDTH / background.width, GAME_HEIGHT / background.height);
      background.setScale(scale);

      const softFocus = this.add.graphics();
      softFocus.fillGradientStyle(0xffffff, 0xffffff, 0xffffff, 0xffffff, 0.18, 0.12, 0.2, 0.2);
      softFocus.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
      return;
    }

    this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x7bdff2);

    const sky = this.add.graphics();
    sky.fillGradientStyle(0x8be7ff, 0x8be7ff, 0xe8fbff, 0xe8fbff, 1);
    sky.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);

    this.add.circle(1128, 94, 58, 0xffe082, 1);
    this.add.circle(1128, 94, 42, 0xfff2ae, 1);

    this.addCloud(154, 88, 1.08);
    this.addCloud(482, 108, 0.82);
    this.addCloud(920, 152, 0.72);

    const hills = this.add.graphics();
    hills.fillStyle(0xb9e97a, 1);
    hills.fillEllipse(250, 724, 760, 260);
    hills.fillEllipse(820, 740, 880, 310);
    hills.fillStyle(0x75d18c, 1);
    hills.fillEllipse(1040, 734, 620, 230);
    hills.fillStyle(0x4fb76f, 1);
    hills.fillRect(0, 646, GAME_WIDTH, 74);

    this.createMarqueeSparkles();
  }

  private createHud(): void {
    this.hud = this.add.container(0, 0);

    if (this.textures.exists(HUD_BAR_TEXTURE)) {
      const bar = this.add.image(640, 60, HUD_BAR_TEXTURE);
      this.hud.add(bar);
    } else {
      const shadow = this.add.graphics();
      shadow.fillStyle(0x2f7f8a, 0.1);
      shadow.fillRoundedRect(72, 34, 1136, 68, 22);
      const bar = this.add.graphics();
      bar.fillStyle(0xf9fdfc, 0.97);
      bar.fillRoundedRect(70, 24, 1140, 68, 22);
      bar.lineStyle(2, 0xc7e8ea, 0.92);
      bar.strokeRoundedRect(70, 24, 1140, 68, 22);
      this.hud.add([shadow, bar]);
    }

    const hudCenterY = 58;

    this.roundText = this.add.text(164, hudCenterY, `1/${this.roundTotal}`, {
      color: '#24546b',
      fontSize: '24px',
      fontStyle: '900',
    });
    this.roundText.setOrigin(0.5);
    this.hud.add(this.roundText);

    const scoreIcon = this.add.image(762, hudCenterY + 2, 'star');
    scoreIcon.setScale(0.42);
    this.hud.add(scoreIcon);
    this.scoreText = this.add.text(776, hudCenterY + 3, '得分 0', {
      color: '#7c4b00',
      fontSize: '19px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 2, right: 2 },
    });
    this.scoreText.setOrigin(0, 0.5);
    this.hud.add(this.scoreText);

    const comboIcon = this.add.image(892, hudCenterY + 2, 'bolt-icon');
    comboIcon.setScale(0.46);
    this.hud.add(comboIcon);
    this.comboText = this.add.text(906, hudCenterY + 3, '连击 0', {
      color: '#a33e62',
      fontSize: '19px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 2, right: 2 },
    });
    this.comboText.setOrigin(0, 0.5);
    this.hud.add(this.comboText);

    this.musicButton = this.createMusicButton(1099, hudCenterY);
    this.hud.add(this.musicButton);
    this.updateMusicButton();

    this.progressDots = [];
    for (let i = 0; i < this.roundTotal; i += 1) {
      const spacing = this.roundTotal > 10 ? 30 : 36;
      const dot = this.add.circle(294 + i * spacing, hudCenterY + 1, 7.2, 0xcfe8ee, 1);
      dot.setName(`progress-${i}`);
      this.progressDots.push(dot);
      this.hud.add(dot);
    }
  }

  private createPlayLayer(): void {
    this.playLayer = this.add.container(0, 0);
    this.overlayLayer = this.add.container(0, 0);

    this.playLayer.add(this.createTaskPanel(650, 220));

    this.mascot = this.createMascot(166, 506);
    this.playLayer.add(this.mascot);

    this.targetBubble = this.createTargetBubble(380, 220, COLORS[0]);
    this.playLayer.add(this.targetBubble);

    this.taskText = this.add.text(520, 220, '找一找 红色', {
      color: '#243f54',
      fontSize: '32px',
      fontStyle: '900',
      stroke: '#f8fffd',
      strokeThickness: 3,
      padding: { top: 12, bottom: 8, left: 8, right: 8 },
    });
    this.taskText.setOrigin(0, 0.5);
    this.playLayer.add(this.taskText);

    this.voiceButton = this.createVoiceButton(874, 222);
    this.playLayer.add(this.voiceButton);

    this.feedbackText = this.add.text(640, 610, '', {
      align: 'center',
      color: '#ffffff',
      fontSize: '34px',
      fontStyle: '800',
      stroke: '#3a6b80',
      strokeThickness: 6,
    });
    this.feedbackText.setOrigin(0.5);
    this.playLayer.add(this.feedbackText);

    this.playLayer.setVisible(false);
    this.hud.setVisible(false);
  }

  private createStartOverlay(): void {
    this.overlayLayer.removeAll(true);

    const shade = this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x24546b, 0.16);
    const card = this.add.graphics();
    card.fillStyle(0xffffff, 0.96);
    card.fillRoundedRect(290, 124, 700, 438, 36);
    card.lineStyle(5, 0xffd447, 1);
    card.strokeRoundedRect(290, 124, 700, 438, 36);

    const title = this.add.text(640, 204, '颜色配对乐园', {
      align: 'center',
      color: '#1f4d64',
      fontSize: '56px',
      fontStyle: '900',
      padding: { top: 14, bottom: 8, left: 8, right: 8 },
    });
    title.setOrigin(0.5);

    const friends = this.add.container(640, 334);
    for (let i = 0; i < 5; i += 1) {
      const color = COLORS[i];
      const tokenKey = this.getTokenTextureKey(color);
      if (this.textures.exists(tokenKey)) {
        const token = this.add.image((i - 2) * 86, 0, tokenKey);
        token.setDisplaySize(74, 74);
        friends.add(token);
      } else {
        const bubble = this.add.circle((i - 2) * 86, 0, 32, color.value, 1);
        const shine = this.add.circle((i - 2) * 86 - 10, -12, 9, 0xffffff, 0.72);
        friends.add([bubble, shine]);
      }
    }

    const startButton = this.createButton(640, 466, 286, 76, '开始游戏', 0xff6b8a, 0xd93f67, () => {
      this.ensureAudio();
      this.unlockVoiceAudio();
      if (this.musicEnabled) {
        this.startBackgroundMusic();
      }
      this.playTone(660, 0.1, 'sine');
      this.startGame();
    });

    this.overlayLayer.add([shade, card, title, friends, startButton]);
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

    this.locked = false;
    this.feedbackText.setText('');
    this.currentTarget = Phaser.Utils.Array.GetRandom(COLORS);
    this.currentPrompt = `prompt-${this.currentTarget.key}`;
    this.taskText.setText(`找一找 ${this.currentTarget.name}`);
    this.roundStartedAt = performance.now();
    this.updateTargetBubble();
    this.updateHud();
    this.createChoices();
    this.time.delayedCall(220, () => this.speakCurrentPrompt());
  }

  private createChoices(): void {
    for (const choice of this.choices) {
      choice.container.destroy();
    }
    this.choices = [];

    const colors = Phaser.Utils.Array.Shuffle([
      this.currentTarget,
      ...Phaser.Utils.Array.Shuffle(COLORS.filter((color) => color.key !== this.currentTarget.key)).slice(
        0,
        this.choiceTotal - 1,
      ),
    ]);

    const choiceSpacing = 164;
    const startX = 692 - ((colors.length - 1) * choiceSpacing) / 2;
    colors.forEach((color, index) => {
      const x = startX + index * choiceSpacing;
      const y = 430 + (index % 2) * 38;
      const bubble = this.createChoiceBubble(x, y, color);
      this.playLayer.add(bubble);
      this.choices.push({ color, container: bubble, baseY: y });
    });
  }

  private createChoiceBubble(x: number, y: number, color: ColorOption): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    container.setSize(142, 156);

    const shadow = this.textures.exists(CHOICE_CUSHION_TEXTURE)
      ? this.add.image(0, 66, CHOICE_CUSHION_TEXTURE)
      : this.add.ellipse(6, 68, 104, 24, 0x24546b, 0.18);
    if (shadow instanceof Phaser.GameObjects.Image) {
      shadow.setDisplaySize(134, 58);
      shadow.setAlpha(0.88);
    }
    const tokenKey = this.getTokenTextureKey(color);
    const tokenObjects: Phaser.GameObjects.GameObject[] = [];
    if (this.textures.exists(tokenKey)) {
      const aura = this.add.circle(0, 0, 67, color.value, 0.08);
      const image = this.add.image(0, -4, tokenKey);
      image.setDisplaySize(124, 124);
      tokenObjects.push(aura, image);
    } else {
      const rope = this.add.graphics();
      rope.lineStyle(4, 0xffffff, 0.72);
      rope.beginPath();
      rope.moveTo(0, 56);
      rope.lineTo(-8, 84);
      rope.lineTo(10, 102);
      rope.strokePath();

      const body = this.add.circle(0, 0, 56, color.value, 1);
      body.setStrokeStyle(5, color.shadow, 0.58);
      const shine = this.add.circle(-18, -22, 14, 0xffffff, 0.78);
      const eyeLeft = this.add.circle(-16, 4, 5, 0x2f4265, 1);
      const eyeRight = this.add.circle(18, 4, 5, 0x2f4265, 1);
      const smile = this.add.graphics();
      smile.lineStyle(4, 0x2f4265, 1);
      smile.beginPath();
      smile.arc(0, 12, 20, Phaser.Math.DegToRad(20), Phaser.Math.DegToRad(160), false);
      smile.strokePath();
      tokenObjects.push(rope, body, shine, eyeLeft, eyeRight, smile);
    }

    const hitZone = this.add.zone(0, 16, 142, 156);
    hitZone.setOrigin(0.5);
    hitZone.setInteractive({ useHandCursor: true });
    hitZone.on('pointerdown', () => this.pickColor(color, container));
    hitZone.on('pointerover', () => {
      this.tweens.add({ targets: container, scale: 1.08, duration: 120, ease: 'Back.Out' });
    });
    hitZone.on('pointerout', () => {
      this.tweens.add({ targets: container, scale: 1, duration: 120, ease: 'Sine.Out' });
    });
    container.add([shadow, ...tokenObjects, hitZone]);

    this.tweens.add({
      targets: container,
      y: y - 12,
      duration: 760 + Phaser.Math.Between(0, 280),
      yoyo: true,
      repeat: -1,
      ease: 'Sine.InOut',
    });

    return container;
  }

  private pickColor(color: ColorOption, bubble: Phaser.GameObjects.Container): void {
    if (this.locked) {
      return;
    }

    this.locked = true;
    const reactionMs = Math.round(performance.now() - this.roundStartedAt);
    const isCorrect = color.key === this.currentTarget.key;
    this.roundEvents.push({
      round: this.roundIndex + 1,
      targetColor: this.currentTarget.key,
      pickedColor: color.key,
      correct: isCorrect,
      reactionMs,
    });

    if (isCorrect) {
      this.correct += 1;
      this.combo += 1;
      this.bestCombo = Math.max(this.bestCombo, this.combo);
      this.feedbackText.setText('');
      this.playVoiceClip(this.combo >= 3 ? 'combo' : 'correct');
      this.playTone(784, 0.08, 'sine');
      this.playTone(1046, 0.12, 'triangle', 0.08);
      this.burstStars(bubble.x, bubble.y);
      this.tweens.add({
        targets: bubble,
        scale: 1.28,
        duration: 150,
        yoyo: true,
        ease: 'Back.Out',
      });
      this.mascotCelebrate();
      this.updateHud();
      this.time.delayedCall(780, () => {
        this.roundIndex += 1;
        this.nextRound();
      });
      return;
    }

    this.wrong += 1;
    this.combo = 0;
    this.feedbackText.setText('');
    const wrongClip = this.playVoiceClip('wrong');
    this.playTone(220, 0.11, 'sawtooth');
    this.tweens.add({
      targets: bubble,
      x: bubble.x + 16,
      scale: 0.94,
      angle: 4,
      duration: 70,
      yoyo: true,
      repeat: 3,
      ease: 'Sine.InOut',
      onComplete: () => {
        bubble.setAngle(0);
        bubble.setScale(1);
      },
    });
    this.showWrongFeedback();
    this.updateHud();
    this.advanceRoundAfterAudio(wrongClip, 1700, 320);
  }

  private showWrongFeedback(): void {
    this.tweens.add({
      targets: [this.targetBubble, this.taskText],
      y: '-=6',
      duration: 90,
      yoyo: true,
      repeat: 1,
      ease: 'Sine.InOut',
    });
    this.tweens.add({
      targets: this.voiceButton,
      scale: 1.06,
      duration: 110,
      yoyo: true,
      ease: 'Back.Out',
    });
    this.tweens.add({
      targets: this.mascot,
      angle: { from: -5, to: 5 },
      duration: 110,
      yoyo: true,
      repeat: 2,
      ease: 'Sine.InOut',
      onComplete: () => this.mascot.setAngle(0),
    });
  }

  private advanceRoundAfterAudio(
    clip: HTMLAudioElement | undefined,
    fallbackMs: number,
    tailTrimMs = 0,
  ): void {
    let advanced = false;
    let cleanup = (): void => {};

    const advance = (): void => {
      if (advanced) {
        return;
      }
      advanced = true;
      cleanup();
      this.roundIndex += 1;
      this.nextRound();
    };

    this.time.delayedCall(fallbackMs, () => {
      advance();
    });

    if (!clip) {
      return;
    }

    const scheduleByDuration = (): void => {
      if (!Number.isFinite(clip.duration) || clip.duration <= 0) {
        return;
      }
      const advanceMs = Math.max(240, clip.duration * 1000 - tailTrimMs);
      this.time.delayedCall(advanceMs, advance);
    };

    const advanceWhenTailStarts = (): void => {
      if (!Number.isFinite(clip.duration) || clip.duration <= 0) {
        return;
      }
      if (clip.currentTime * 1000 >= clip.duration * 1000 - tailTrimMs) {
        advance();
      }
    };

    cleanup = (): void => {
      clip.removeEventListener('timeupdate', advanceWhenTailStarts);
      clip.removeEventListener('loadedmetadata', scheduleByDuration);
    };
    scheduleByDuration();
    clip.addEventListener('loadedmetadata', scheduleByDuration, { once: true });
    clip.addEventListener('timeupdate', advanceWhenTailStarts);
    clip.addEventListener(
      'ended',
      () => {
        advance();
      },
      { once: true },
    );
  }

  private finishGame(): void {
    const endedAt = performance.now();
    const durationMs = Math.round(endedAt - this.startedAt);
    const avgReactionMs =
      this.roundEvents.length === 0
        ? 0
        : Math.round(
            this.roundEvents.reduce((sum, event) => sum + event.reactionMs, 0) / this.roundEvents.length,
          );
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

    const shade = this.add.rectangle(GAME_WIDTH / 2, GAME_HEIGHT / 2, GAME_WIDTH, GAME_HEIGHT, 0x24546b, 0.28);
    const card = this.add.graphics();
    card.fillStyle(0xffffff, 0.97);
    card.fillRoundedRect(264, 82, 752, 558, 38);
    card.lineStyle(6, 0x54ce73, 1);
    card.strokeRoundedRect(264, 82, 752, 558, 38);

    const title = this.add.text(640, 150, result.stars >= 2 ? '闯关成功！' : '完成练习！', {
      align: 'center',
      color: '#24546b',
      fontSize: '52px',
      fontStyle: '900',
      padding: { top: 14, bottom: 8, left: 8, right: 8 },
    });
    title.setOrigin(0.5);

    const stars = this.add.container(640, 240);
    for (let i = 0; i < 3; i += 1) {
      const star = this.add.image((i - 1) * 86, 0, 'star');
      star.setScale(i < result.stars ? 1.6 : 1.2);
      star.setAlpha(i < result.stars ? 1 : 0.28);
      stars.add(star);
      if (i < result.stars) {
        this.tweens.add({
          targets: star,
          y: -12,
          duration: 420 + i * 120,
          yoyo: true,
          repeat: -1,
          ease: 'Sine.InOut',
        });
      }
    }

    const stats = this.add.text(
      640,
      338,
      `答对 ${result.correct} / ${result.total}    正确率 ${Math.round(result.accuracy * 100)}%\n最佳连击 ${result.bestCombo}    平均反应 ${Math.round(result.avgReactionMs / 100) / 10} 秒`,
      {
        align: 'center',
        color: '#41596b',
        fontSize: '28px',
        fontStyle: '800',
        lineSpacing: 16,
        padding: { top: 10, bottom: 8, left: 8, right: 8 },
      },
    );
    stats.setOrigin(0.5);

    const replayButton = this.createButton(514, 512, 210, 68, '再玩一次', 0x4e9cff, 0x2767c7, () => {
      this.playTone(660, 0.1, 'sine');
      this.startGame();
    });
    const doneButton = this.createButton(766, 512, 210, 68, '完成', 0xff9b42, 0xd56517, () => {
      requestClose(this.lastResult);
    });

    this.overlayLayer.add([shade, card, title, stars, stats, replayButton, doneButton]);
  }

  private updateHud(): void {
    this.roundText.setText(`${Math.min(this.roundIndex + 1, this.roundTotal)}/${this.roundTotal}`);
    this.scoreText.setText(`得分 ${this.correct}`);
    this.comboText.setText(`连击 ${this.combo}`);

    this.progressDots.forEach((dot, index) => {
      dot.setFillStyle(index < this.roundIndex ? 0x55c98f : index === this.roundIndex ? 0xffcf5a : 0xcfe8ee);
      dot.setScale(index === this.roundIndex ? 1.24 : 1);
    });
  }

  private updateTargetBubble(): void {
    this.targetBubble.destroy();
    this.targetBubble = this.createTargetBubble(380, 220, this.currentTarget);
    this.playLayer.add(this.targetBubble);
  }

  private createVoiceButton(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const width = 150;
    const height = 50;
    const button = this.textures.exists(LISTEN_BUTTON_TEXTURE)
      ? this.add.image(0, 0, LISTEN_BUTTON_TEXTURE)
      : undefined;

    const label = this.add.text(-10, 1, '再听', {
      color: '#24546b',
      fontSize: '22px',
      fontStyle: '900',
      padding: { top: 8, bottom: 6, left: 4, right: 4 },
    });
    label.setOrigin(0, 0.5);

    const hitZone = this.add.zone(0, 3, width, height + 14);
    hitZone.setOrigin(0.5);
    hitZone.setInteractive({ useHandCursor: true });
    hitZone.on('pointerdown', () => {
      this.animateVoiceButton();
      this.speakCurrentPrompt();
    });
    hitZone.on('pointerover', () => this.tweens.add({ targets: container, scale: 1.06, duration: 120 }));
    hitZone.on('pointerout', () => this.tweens.add({ targets: container, scale: 1, duration: 120 }));

    if (button) {
      container.add([button, label, hitZone]);
    } else {
      const body = this.add.graphics();
      body.fillStyle(0xf9fdfc, 0.98);
      body.fillRoundedRect(-width / 2, -height / 2, width, height, 16);
      body.lineStyle(2, 0xc7e8ea, 1);
      body.strokeRoundedRect(-width / 2, -height / 2, width, height, 16);
      container.add([body, label, hitZone]);
    }

    return container;
  }

  private createTaskPanel(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);

    if (this.textures.exists(TASK_CARD_TEXTURE)) {
      const card = this.add.image(0, 0, TASK_CARD_TEXTURE);
      container.add(card);
      return container;
    }

    const shadow = this.add.graphics();
    shadow.fillStyle(0x2f7f8a, 0.1);
    shadow.fillRoundedRect(-326, -50, 652, 104, 30);
    const panel = this.add.graphics();
    panel.fillStyle(0xf9fdfc, 0.96);
    panel.fillRoundedRect(-330, -56, 660, 104, 30);
    panel.lineStyle(2, 0xc7e8ea, 1);
    panel.strokeRoundedRect(-330, -56, 660, 104, 30);

    const leftWell = this.add.circle(-212, -4, 56, 0xffffff, 0.76);
    leftWell.setStrokeStyle(2, 0xd5ecee, 1);
    const textDivider = this.add.rectangle(-98, -4, 2, 56, 0xdceff0, 1);
    const buttonWell = this.add.graphics();
    buttonWell.fillStyle(0xedf7f7, 0.82);
    buttonWell.fillRoundedRect(110, -31, 176, 62, 18);
    buttonWell.lineStyle(2, 0xd7eef0, 0.86);
    buttonWell.strokeRoundedRect(110, -31, 176, 62, 18);

    container.add([shadow, panel, leftWell, textDivider, buttonWell]);
    return container;
  }

  private createMusicButton(x: number, y: number): Phaser.GameObjects.Container {
    const width = 138;
    const height = 54;
    const container = this.add.container(x, y);
    const body = this.add.image(0, 0, this.musicEnabled ? MUSIC_TOGGLE_ON_TEXTURE : MUSIC_TOGGLE_OFF_TEXTURE);
    body.setName('music-toggle-image');

    const hitZone = this.add.zone(0, 2, width, height + 8);
    hitZone.setOrigin(0.5);
    hitZone.setInteractive({ useHandCursor: true });
    hitZone.on('pointerdown', () => this.toggleBackgroundMusic());
    hitZone.on('pointerover', () => this.tweens.add({ targets: container, scale: 1.05, duration: 120 }));
    hitZone.on('pointerout', () => this.tweens.add({ targets: container, scale: 1, duration: 120 }));

    container.add([body, hitZone]);
    return container;
  }

  private updateMusicButton(): void {
    if (!this.musicButton) {
      return;
    }

    const body = this.musicButton.getByName('music-toggle-image') as Phaser.GameObjects.Image | null;
    body?.setTexture(this.musicEnabled ? MUSIC_TOGGLE_ON_TEXTURE : MUSIC_TOGGLE_OFF_TEXTURE);
  }

  private animateVoiceButton(): void {
    this.tweens.add({
      targets: this.voiceButton,
      scale: 1.12,
      duration: 110,
      yoyo: true,
      ease: 'Back.Out',
    });
  }

  private createTargetBubble(x: number, y: number, color: ColorOption): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const glow = this.add.circle(0, 0, 76, color.value, 0.22);
    const tokenKey = this.getTokenTextureKey(color);
    if (this.textures.exists(tokenKey)) {
      const ring = this.add.circle(0, 0, 61, 0xffffff, 0.9);
      ring.setStrokeStyle(5, color.shadow, 0.5);
      const image = this.add.image(0, -2, tokenKey);
      image.setDisplaySize(108, 108);
      container.add([glow, ring, image]);
    } else {
      const body = this.add.circle(0, 0, 52, color.value, 1);
      body.setStrokeStyle(5, color.shadow, 0.68);
      const shine = this.add.circle(-18, -20, 13, 0xffffff, 0.75);
      container.add([glow, body, shine]);
    }
    this.tweens.add({ targets: container, scale: 1.08, duration: 620, yoyo: true, repeat: -1 });
    return container;
  }

  private createMascot(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const shadow = this.add.ellipse(4, 82, 116, 26, 0x24546b, 0.15);
    if (this.textures.exists(MASCOT_TEXTURE)) {
      const glow = this.add.circle(0, -8, 88, 0xffffff, 0.28);
      const mascot = this.add.image(0, -8, MASCOT_TEXTURE);
      mascot.setDisplaySize(178, 178);
      container.add([shadow, glow, mascot]);
      this.tweens.add({ targets: container, y: y - 10, duration: 880, yoyo: true, repeat: -1, ease: 'Sine.InOut' });
      return container;
    }

    const glow = this.add.image(0, 0, 'star-white');
    glow.setScale(2.65);
    glow.setAlpha(0.34);
    const body = this.add.image(0, 0, 'star');
    body.setScale(2.15);
    const eyeLeft = this.add.circle(-20, -8, 6, 0x24546b, 1);
    const eyeRight = this.add.circle(20, -8, 6, 0x24546b, 1);
    const cheekLeft = this.add.circle(-32, 14, 9, 0xff8fab, 0.72);
    const cheekRight = this.add.circle(32, 14, 9, 0xff8fab, 0.72);
    const mouth = this.add.graphics();
    mouth.lineStyle(4, 0x24546b, 1);
    mouth.beginPath();
    mouth.arc(0, 0, 21, Phaser.Math.DegToRad(28), Phaser.Math.DegToRad(152), false);
    mouth.strokePath();
    container.add([
      shadow,
      glow,
      body,
      eyeLeft,
      eyeRight,
      cheekLeft,
      cheekRight,
      mouth,
    ]);
    this.tweens.add({ targets: container, y: y - 10, duration: 880, yoyo: true, repeat: -1, ease: 'Sine.InOut' });
    return container;
  }

  private getTokenTextureKey(color: ColorOption): string {
    return `${TOKEN_TEXTURE_PREFIX}-${color.key}`;
  }

  private mascotCelebrate(): void {
    this.tweens.add({
      targets: this.mascot,
      angle: { from: -8, to: 8 },
      scale: 1.08,
      duration: 120,
      yoyo: true,
      repeat: 3,
      ease: 'Sine.InOut',
      onComplete: () => {
        this.mascot.setAngle(0);
        this.mascot.setScale(1);
      },
    });
  }

  private burstStars(x: number, y: number): void {
    for (let i = 0; i < 18; i += 1) {
      const star = this.add.image(x, y, i % 3 === 0 ? 'star-white' : 'star');
      star.setScale(Phaser.Math.FloatBetween(0.45, 0.9));
      this.playLayer.add(star);
      this.tweens.add({
        targets: star,
        x: x + Phaser.Math.Between(-120, 120),
        y: y + Phaser.Math.Between(-110, 70),
        angle: Phaser.Math.Between(-240, 240),
        alpha: 0,
        duration: Phaser.Math.Between(520, 820),
        ease: 'Cubic.Out',
        onComplete: () => star.destroy(),
      });
    }
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
    shadow.fillRoundedRect(-width / 2, -height / 2 + 8, width, height, 24);
    const body = this.add.graphics();
    body.fillStyle(color, 1);
    body.fillRoundedRect(-width / 2, -height / 2, width, height, 24);
    body.lineStyle(4, 0xffffff, 0.55);
    body.strokeRoundedRect(-width / 2, -height / 2, width, height, 24);
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
    hitZone.on('pointerdown', () => {
      this.tweens.add({ targets: container, scale: 0.94, duration: 70, yoyo: true, onComplete: onClick });
    });
    hitZone.on('pointerover', () => this.tweens.add({ targets: container, scale: 1.04, duration: 120 }));
    hitZone.on('pointerout', () => this.tweens.add({ targets: container, scale: 1, duration: 120 }));
    container.add([shadow, body, text, hitZone]);
    container.setSize(width, height + 8);
    return container;
  }

  private addCloud(x: number, y: number, scale: number): void {
    const cloud = this.add.container(x, y);
    cloud.setScale(scale);
    cloud.add([
      this.add.circle(-48, 12, 28, 0xffffff, 0.82),
      this.add.circle(-12, -4, 40, 0xffffff, 0.9),
      this.add.circle(34, 10, 30, 0xffffff, 0.84),
      this.add.rectangle(-4, 22, 112, 36, 0xffffff, 0.84),
    ]);
    this.tweens.add({
      targets: cloud,
      x: x + 28,
      duration: 5600 + x * 4,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.InOut',
    });
  }

  private createMarqueeSparkles(): void {
    const layer = this.add.container(0, 0);
    const colors = [0xffffff, 0xfff0a8, 0xcff7ff, 0xffffff];

    for (let i = 0; i < 36; i += 1) {
      const x = 28 + i * 35;
      const y = 638 + Math.sin(i * 1.42) * 18;
      const color = colors[i % colors.length];
      const sparkle = this.add.container(x, y);
      const glow = this.add.circle(0, 0, 16, color, 0.16);
      const core = this.add.circle(0, 0, 6, color, 0.72);
      const flareH = this.add.rectangle(0, 0, 26, 3, color, 0.42);
      const flareV = this.add.rectangle(0, 0, 3, 26, color, 0.42);
      sparkle.add([glow, flareH, flareV, core]);
      sparkle.setAlpha(0.36);
      layer.add(sparkle);

      this.tweens.add({
        targets: sparkle,
        x: x + 24,
        y: y - 3,
        alpha: 1,
        scale: 1.22,
        angle: 90,
        duration: 780,
        delay: i * 90,
        yoyo: true,
        repeat: -1,
        repeatDelay: 1750,
        ease: 'Sine.InOut',
      });
    }
  }

  private makeCircleTexture(key: string, radius: number, color: number): void {
    const graphics = this.add.graphics();
    graphics.fillStyle(color, 1);
    graphics.fillCircle(radius, radius, radius);
    graphics.generateTexture(key, radius * 2, radius * 2);
    graphics.destroy();
  }

  private makeStarTexture(key: string, fill: number, stroke: number): void {
    const graphics = this.add.graphics();
    const points: Phaser.Math.Vector2[] = [];
    for (let i = 0; i < 10; i += 1) {
      const radius = i % 2 === 0 ? 30 : 13;
      const angle = Phaser.Math.DegToRad(i * 36 - 90);
      points.push(new Phaser.Math.Vector2(32 + Math.cos(angle) * radius, 32 + Math.sin(angle) * radius));
    }
    graphics.fillStyle(fill, 1);
    graphics.lineStyle(4, stroke, 1);
    graphics.beginPath();
    graphics.moveTo(points[0].x, points[0].y);
    for (const point of points.slice(1)) {
      graphics.lineTo(point.x, point.y);
    }
    graphics.closePath();
    graphics.fillPath();
    graphics.strokePath();
    graphics.generateTexture(key, 64, 64);
    graphics.destroy();
  }

  private makeBoltTexture(key: string, fill: number, stroke: number): void {
    const graphics = this.add.graphics();
    graphics.fillStyle(fill, 1);
    graphics.lineStyle(4, stroke, 1);
    graphics.beginPath();
    graphics.moveTo(34, 4);
    graphics.lineTo(12, 38);
    graphics.lineTo(33, 38);
    graphics.lineTo(22, 70);
    graphics.lineTo(58, 28);
    graphics.lineTo(37, 28);
    graphics.closePath();
    graphics.fillPath();
    graphics.strokePath();
    graphics.generateTexture(key, 72, 76);
    graphics.destroy();
  }

  private makeMusicTexture(key: string, fill: number, stroke: number): void {
    const graphics = this.add.graphics();
    graphics.fillStyle(fill, 1);
    graphics.lineStyle(4, stroke, 1);
    graphics.fillRoundedRect(18, 22, 10, 34, 4);
    graphics.fillRoundedRect(46, 12, 10, 38, 4);
    graphics.fillCircle(20, 58, 13);
    graphics.fillCircle(48, 52, 13);
    graphics.fillTriangle(28, 18, 56, 8, 56, 20);
    graphics.lineStyle(4, stroke, 0.9);
    graphics.strokeCircle(20, 58, 13);
    graphics.strokeCircle(48, 52, 13);

    if (key.includes('muted')) {
      graphics.lineStyle(6, 0xffffff, 1);
      graphics.beginPath();
      graphics.moveTo(12, 14);
      graphics.lineTo(62, 66);
      graphics.strokePath();
      graphics.lineStyle(3, 0x7c8b96, 1);
      graphics.beginPath();
      graphics.moveTo(12, 14);
      graphics.lineTo(62, 66);
      graphics.strokePath();
    }

    graphics.generateTexture(key, 76, 76);
    graphics.destroy();
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

  private playTone(
    frequency: number,
    duration: number,
    type: OscillatorType,
    delay = 0,
  ): void {
    this.ensureAudio();
    if (!this.audioContext) {
      return;
    }

    const start = this.audioContext.currentTime + delay;
    const oscillator = this.audioContext.createOscillator();
    const gain = this.audioContext.createGain();
    oscillator.type = type;
    oscillator.frequency.value = frequency;
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(0.12, start + 0.015);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + duration);
    oscillator.connect(gain);
    gain.connect(this.audioContext.destination);
    oscillator.start(start);
    oscillator.stop(start + duration + 0.03);
  }

  private toggleBackgroundMusic(): void {
    this.ensureAudio();
    this.musicEnabled = !this.musicEnabled;
    window.localStorage.setItem('colorMatchMusic', this.musicEnabled ? 'on' : 'off');
    this.updateMusicButton();

    if (this.musicEnabled) {
      this.startBackgroundMusic();
      this.playTone(880, 0.08, 'sine');
    } else {
      this.stopBackgroundMusic();
    }
  }

  private startBackgroundMusic(): void {
    this.ensureAudio();
    if (!this.audioContext || this.musicTimer !== undefined) {
      return;
    }

    if (!this.musicGain) {
      this.musicGain = this.audioContext.createGain();
      this.musicGain.gain.setValueAtTime(0.0001, this.audioContext.currentTime);
      this.musicGain.connect(this.audioContext.destination);
    }

    this.musicGain.gain.cancelScheduledValues(this.audioContext.currentTime);
    this.musicGain.gain.exponentialRampToValueAtTime(0.5, this.audioContext.currentTime + 0.6);
    this.scheduleMusicLoop();
  }

  private stopBackgroundMusic(): void {
    if (this.musicTimer !== undefined) {
      window.clearTimeout(this.musicTimer);
      this.musicTimer = undefined;
    }

    if (this.musicGain && this.audioContext) {
      this.musicGain.gain.cancelScheduledValues(this.audioContext.currentTime);
      this.musicGain.gain.exponentialRampToValueAtTime(0.0001, this.audioContext.currentTime + 0.35);
    }
  }

  private stopAllAudio(): void {
    this.stopBackgroundMusic();
    this.stopVoiceClip();

    if (this.audioContext?.state === 'running') {
      void this.audioContext.suspend();
    }
  }

  private scheduleMusicLoop(): void {
    if (!this.audioContext || !this.musicGain || !this.musicEnabled) {
      this.musicTimer = undefined;
      return;
    }

    const melody = [
      523.25,
      659.25,
      783.99,
      659.25,
      587.33,
      698.46,
      880,
      698.46,
    ];
    const start = this.audioContext.currentTime + 0.04;
    melody.forEach((frequency, index) => {
      this.playMusicNote(frequency, start + index * 0.24, 0.18);
    });

    this.musicTimer = window.setTimeout(() => {
      this.musicTimer = undefined;
      this.scheduleMusicLoop();
    }, 2200);
  }

  private playMusicNote(frequency: number, start: number, duration: number): void {
    if (!this.audioContext || !this.musicGain) {
      return;
    }

    const oscillator = this.audioContext.createOscillator();
    const gain = this.audioContext.createGain();
    oscillator.type = 'triangle';
    oscillator.frequency.setValueAtTime(frequency, start);
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(0.24, start + 0.025);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + duration);
    oscillator.connect(gain);
    gain.connect(this.musicGain);
    oscillator.start(start);
    oscillator.stop(start + duration + 0.04);
  }

  private speakCurrentPrompt(): void {
    this.animateVoiceButton();
    this.playVoiceClip(this.currentPrompt);
  }

  private playVoiceClip(key: string): HTMLAudioElement | undefined {
    const clip = this.getVoiceClip(key);
    if (!clip) {
      return undefined;
    }

    this.stopVoiceClip();
    clip.currentTime = 0;
    clip.volume = 1;
    this.activeVoice = clip;
    void clip.play().catch(() => {
      this.activeVoice = undefined;
    });
    return clip;
  }

  private stopVoiceClip(): void {
    if (!this.activeVoice) {
      return;
    }

    this.activeVoice.pause();
    this.activeVoice.currentTime = 0;
    this.activeVoice = undefined;
  }

  private getVoiceClip(key: string): HTMLAudioElement | undefined {
    if (this.voiceClips.has(key)) {
      return this.voiceClips.get(key);
    }

    const clip = new Audio(publicAssetPath(`audio/color-match/${key}.mp3`));
    clip.preload = 'auto';
    this.voiceClips.set(key, clip);
    return clip;
  }

  private unlockVoiceAudio(): void {
    const clip = this.getVoiceClip('correct');
    if (!clip) {
      return;
    }

    clip.muted = true;
    void clip
      .play()
      .then(() => {
        clip.pause();
        clip.currentTime = 0;
        clip.muted = false;
      })
      .catch(() => {
        clip.muted = false;
      });
  }
}
