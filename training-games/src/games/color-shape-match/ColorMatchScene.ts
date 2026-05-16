import Phaser from 'phaser';
import {
  GameLaunchParams,
  GameResult,
  requestClose,
  submitGameResult,
} from '../../platform/hostBridge';

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

  private hud!: Phaser.GameObjects.Container;
  private playLayer!: Phaser.GameObjects.Container;
  private overlayLayer!: Phaser.GameObjects.Container;
  private roundText!: Phaser.GameObjects.Text;
  private scoreText!: Phaser.GameObjects.Text;
  private comboText!: Phaser.GameObjects.Text;
  private musicButton!: Phaser.GameObjects.Container;
  private musicButtonText!: Phaser.GameObjects.Text;
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

  create(): void {
    this.musicEnabled = window.localStorage.getItem('colorMatchMusic') !== 'off';
    this.createGeneratedTextures();
    this.createWorld();
    this.createHud();
    this.createPlayLayer();
    this.createStartOverlay();
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

    const bar = this.add.graphics();
    bar.fillStyle(0xffffff, 0.88);
    bar.fillRoundedRect(36, 24, 1208, 86, 28);
    bar.lineStyle(4, 0x4fc3dc, 0.35);
    bar.strokeRoundedRect(36, 24, 1208, 86, 28);
    this.hud.add(bar);

    const hudCenterY = 67;

    const progressBadge = this.add.graphics();
    progressBadge.fillStyle(0xffffff, 0.82);
    progressBadge.fillRoundedRect(70, 40, 104, 54, 18);
    progressBadge.lineStyle(3, 0xb8dff1, 0.64);
    progressBadge.strokeRoundedRect(70, 40, 104, 54, 18);
    this.hud.add(progressBadge);

    this.roundText = this.add.text(122, hudCenterY, `1/${this.roundTotal}`, {
      color: '#24546b',
      fontSize: '24px',
      fontStyle: '800',
    });
    this.roundText.setOrigin(0.5);
    this.hud.add(this.roundText);

    const scoreIcon = this.add.image(920, hudCenterY, 'star');
    scoreIcon.setScale(0.62);
    this.hud.add(scoreIcon);
    this.scoreText = this.add.text(954, hudCenterY, '0', {
      color: '#7c4b00',
      fontSize: '24px',
      fontStyle: '800',
    });
    this.scoreText.setOrigin(0, 0.5);
    this.hud.add(this.scoreText);

    const comboIcon = this.add.image(1066, hudCenterY, 'bolt-icon');
    comboIcon.setScale(0.72);
    this.hud.add(comboIcon);
    this.comboText = this.add.text(1102, hudCenterY, '0', {
      color: '#a33e62',
      fontSize: '24px',
      fontStyle: '800',
    });
    this.comboText.setOrigin(0, 0.5);
    this.hud.add(this.comboText);

    this.progressDots = [];
    for (let i = 0; i < this.roundTotal; i += 1) {
      const dot = this.add.circle(278 + i * 32, hudCenterY, 8, 0xb8dff1, 1);
      dot.setName(`progress-${i}`);
      this.progressDots.push(dot);
      this.hud.add(dot);
    }
  }

  private createPlayLayer(): void {
    this.playLayer = this.add.container(0, 0);
    this.overlayLayer = this.add.container(0, 0);

    this.musicButton = this.createMusicButton(1150, 144);
    this.playLayer.add(this.musicButton);
    this.updateMusicButton();

    this.mascot = this.createMascot(166, 506);
    this.playLayer.add(this.mascot);

    this.targetBubble = this.createTargetBubble(430, 218, COLORS[0]);
    this.playLayer.add(this.targetBubble);

    this.taskText = this.add.text(548, 218, '找 红色', {
      color: '#2f4265',
      fontSize: '38px',
      fontStyle: '900',
      stroke: '#ffffff',
      strokeThickness: 8,
      padding: { top: 12, bottom: 8, left: 8, right: 8 },
    });
    this.taskText.setOrigin(0, 0.5);
    this.playLayer.add(this.taskText);

    this.voiceButton = this.createVoiceButton(820, 218);
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
      const bubble = this.add.circle((i - 2) * 86, 0, 32, color.value, 1);
      const shine = this.add.circle((i - 2) * 86 - 10, -12, 9, 0xffffff, 0.72);
      friends.add([bubble, shine]);
    }

    const startButton = this.createButton(640, 466, 286, 76, '开始游戏', 0xff6b8a, 0xd93f67, () => {
      this.ensureAudio();
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
    this.currentPrompt = `请找到${this.currentTarget.name}气球`;
    this.taskText.setText(`找 ${this.currentTarget.name}`);
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
    const startX = 640 - ((colors.length - 1) * choiceSpacing) / 2;
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

    const shadow = this.add.ellipse(6, 68, 104, 24, 0x24546b, 0.18);
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
    container.add([shadow, rope, body, shine, eyeLeft, eyeRight, smile, hitZone]);

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
      this.speak(this.combo >= 3 ? `太棒啦，${this.combo}连击` : '找对啦');
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
    } else {
      this.wrong += 1;
      this.combo = 0;
      this.feedbackText.setText('');
      this.speak(`再看看，请找${this.currentTarget.name}气球`);
      this.playTone(220, 0.11, 'sawtooth');
      this.cameras.main.shake(120, 0.004);
      this.tweens.add({
        targets: bubble,
        x: bubble.x + 16,
        duration: 55,
        yoyo: true,
        repeat: 3,
      });
    }

    this.updateHud();
    this.time.delayedCall(isCorrect ? 780 : 980, () => {
      this.roundIndex += 1;
      this.nextRound();
    });
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
    this.scoreText.setText(`${this.correct}`);
    this.comboText.setText(`${this.combo}`);

    this.progressDots.forEach((dot, index) => {
      dot.setFillStyle(index < this.roundIndex ? 0x54ce73 : index === this.roundIndex ? 0xffd447 : 0xb8dff1);
      dot.setScale(index === this.roundIndex ? 1.3 : 1);
    });
  }

  private updateTargetBubble(): void {
    this.targetBubble.destroy();
    this.targetBubble = this.createTargetBubble(430, 218, this.currentTarget);
    this.playLayer.addAt(this.targetBubble, 1);
  }

  private createVoiceButton(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const width = 200;
    const height = 64;
    const shadow = this.add.graphics();
    shadow.fillStyle(0x49a9c5, 0.92);
    shadow.fillRoundedRect(-width / 2, -height / 2 + 7, width, height, 24);
    const body = this.add.graphics();
    body.fillStyle(0xffffff, 0.96);
    body.fillRoundedRect(-width / 2, -height / 2, width, height, 24);
    body.lineStyle(4, 0x4fc3dc, 0.82);
    body.strokeRoundedRect(-width / 2, -height / 2, width, height, 24);

    const iconBg = this.add.circle(-48, -1, 22, 0x4e9cff, 1);
    iconBg.setStrokeStyle(3, 0xffffff, 0.76);

    const speaker = this.add.graphics();
    speaker.fillStyle(0xffffff, 1);
    speaker.fillRoundedRect(-58, -11, 9, 20, 3);
    speaker.fillTriangle(-49, -15, -49, 13, -33, -1);
    speaker.lineStyle(4, 0xffffff, 1);
    speaker.beginPath();
    speaker.arc(-29, -1, 9, Phaser.Math.DegToRad(-38), Phaser.Math.DegToRad(38), false);
    speaker.strokePath();
    speaker.beginPath();
    speaker.arc(-25, -1, 16, Phaser.Math.DegToRad(-34), Phaser.Math.DegToRad(34), false);
    speaker.strokePath();

    const label = this.add.text(-22, 1, '再听一次', {
      color: '#24546b',
      fontSize: '25px',
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

    container.add([shadow, body, iconBg, speaker, label, hitZone]);

    return container;
  }

  private createMusicButton(x: number, y: number): Phaser.GameObjects.Container {
    const width = 106;
    const height = 52;
    const container = this.add.container(x, y);
    const shadow = this.add.graphics();
    shadow.fillStyle(0x49a9c5, 0.65);
    shadow.fillRoundedRect(-width / 2, -height / 2 + 5, width, height, 18);
    const body = this.add.graphics();
    body.fillStyle(0xffffff, 0.92);
    body.fillRoundedRect(-width / 2, -height / 2, width, height, 18);
    body.lineStyle(3, 0xb8dff1, 0.88);
    body.strokeRoundedRect(-width / 2, -height / 2, width, height, 18);

    const icon = this.add.image(-24, 0, 'music-icon');
    icon.setName('music-icon');
    icon.setScale(0.56);
    this.musicButtonText = this.add.text(4, 1, '开', {
      color: '#24546b',
      fontSize: '22px',
      fontStyle: '900',
      padding: { top: 6, bottom: 4, left: 2, right: 2 },
    });
    this.musicButtonText.setOrigin(0, 0.5);

    const hitZone = this.add.zone(0, 2, width, height + 10);
    hitZone.setOrigin(0.5);
    hitZone.setInteractive({ useHandCursor: true });
    hitZone.on('pointerdown', () => this.toggleBackgroundMusic());
    hitZone.on('pointerover', () => this.tweens.add({ targets: container, scale: 1.05, duration: 120 }));
    hitZone.on('pointerout', () => this.tweens.add({ targets: container, scale: 1, duration: 120 }));

    container.add([shadow, body, icon, this.musicButtonText, hitZone]);
    return container;
  }

  private updateMusicButton(): void {
    if (!this.musicButton || !this.musicButtonText) {
      return;
    }

    const icon = this.musicButton.getByName('music-icon') as Phaser.GameObjects.Image | null;
    icon?.setTexture(this.musicEnabled ? 'music-icon' : 'music-muted-icon');
    this.musicButtonText.setText(this.musicEnabled ? '开' : '关');
    this.musicButtonText.setColor(this.musicEnabled ? '#24546b' : '#7c8b96');
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
    const body = this.add.circle(0, 0, 52, color.value, 1);
    body.setStrokeStyle(5, color.shadow, 0.68);
    const shine = this.add.circle(-18, -20, 13, 0xffffff, 0.75);
    container.add([glow, body, shine]);
    this.tweens.add({ targets: container, scale: 1.08, duration: 620, yoyo: true, repeat: -1 });
    return container;
  }

  private createMascot(x: number, y: number): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);
    const shadow = this.add.ellipse(4, 82, 116, 26, 0x24546b, 0.15);
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
    this.speak(this.currentPrompt);
  }

  private speak(text: string): void {
    const synthesis = window.speechSynthesis;
    if (!synthesis || !text) {
      return;
    }

    synthesis.cancel();
    const utterance = new SpeechSynthesisUtterance(text);
    utterance.lang = 'zh-CN';
    utterance.rate = 0.9;
    utterance.pitch = 1.18;
    utterance.volume = 1;

    const voices = synthesis.getVoices();
    const zhVoice = voices.find((voice) => voice.lang.toLowerCase().startsWith('zh'));
    if (zhVoice) {
      utterance.voice = zhVoice;
    }

    try {
      synthesis.speak(utterance);
    } catch {
      // Some WebView containers block speech synthesis; sound effects still run.
    }
  }
}
