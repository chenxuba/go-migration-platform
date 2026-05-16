export type GameHost = 'flutter' | 'uniapp' | 'browser';

export interface GameLaunchParams {
  gameId: string;
  taskId: string;
  studentId: string;
  token: string;
  host: GameHost;
  difficulty: 'easy' | 'normal' | 'hard';
}

export interface GameResult {
  gameId: string;
  taskId: string;
  studentId: string;
  startedAt: string;
  endedAt: string;
  durationMs: number;
  total: number;
  correct: number;
  wrong: number;
  accuracy: number;
  bestCombo: number;
  avgReactionMs: number;
  stars: number;
  events: Array<{
    round: number;
    targetColor: string;
    pickedColor: string;
    correct: boolean;
    reactionMs: number;
  }>;
}

declare global {
  interface Window {
    FlutterTrainingGame?: {
      postMessage: (message: string) => void;
    };
    uni?: {
      postMessage: (options: { data: unknown }) => void;
    };
    wx?: {
      miniProgram?: {
        postMessage: (options: { data: unknown }) => void;
        navigateBack?: (options?: { delta?: number }) => void;
      };
    };
  }
}

export function readLaunchParams(): GameLaunchParams {
  const params = new URLSearchParams(window.location.search);

  return {
    gameId: params.get('gameId') || 'color-shape-match',
    taskId: params.get('taskId') || 'demo-task',
    studentId: params.get('studentId') || 'demo-student',
    token: params.get('token') || '',
    host: parseHost(params.get('host')),
    difficulty: parseDifficulty(params.get('difficulty')),
  };
}

export async function submitGameResult(result: GameResult, token: string): Promise<void> {
  postHostMessage('training-game-result', result);

  if (!token) {
    return;
  }

  try {
    await fetch('/api/training-game/results', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(result),
    });
  } catch {
    postHostMessage('training-game-submit-failed', result);
  }
}

export function requestClose(result?: GameResult): void {
  postHostMessage('training-game-close', result || null);

  if (window.wx?.miniProgram?.navigateBack) {
    window.wx.miniProgram.navigateBack({ delta: 1 });
  }
}

function postHostMessage(type: string, payload: unknown): void {
  const message = { type, payload };

  window.parent?.postMessage(message, '*');

  if (window.FlutterTrainingGame) {
    window.FlutterTrainingGame.postMessage(JSON.stringify(message));
  }

  if (window.uni) {
    window.uni.postMessage({ data: message });
  }

  if (window.wx?.miniProgram) {
    window.wx.miniProgram.postMessage({ data: message });
  }
}

function parseHost(value: string | null): GameHost {
  if (value === 'flutter' || value === 'uniapp') {
    return value;
  }

  return 'browser';
}

function parseDifficulty(value: string | null): GameLaunchParams['difficulty'] {
  if (value === 'easy' || value === 'hard') {
    return value;
  }

  return 'normal';
}
