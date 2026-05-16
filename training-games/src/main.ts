import Phaser from 'phaser';
import './style.css';
import { ColorMatchScene } from './games/color-shape-match/ColorMatchScene';
import { readLaunchParams } from './platform/hostBridge';

const launchParams = readLaunchParams();

new Phaser.Game({
  type: Phaser.AUTO,
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
  scene: [new ColorMatchScene(launchParams)],
});
