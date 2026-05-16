# 颜色配对乐园

面向 Pad App WebView 和 UniApp 小程序 `web-view` 的 H5 儿童训练游戏 MVP。

## 本地运行

```bash
npm install
npm run dev
```

默认地址：

```text
http://localhost:6677/
```

带宿主参数示例：

```text
http://localhost:6677/?host=uniapp&taskId=demo-task&studentId=demo-student&difficulty=normal
```

## 游戏结果

游戏结束后会向宿主发送 `training-game-result` 消息，并在传入 `token` 时尝试提交：

```text
POST /api/training-game/results
```

结果字段包括：

- `gameId`
- `taskId`
- `studentId`
- `durationMs`
- `total`
- `correct`
- `wrong`
- `accuracy`
- `bestCombo`
- `avgReactionMs`
- `stars`
- `events`

## 宿主通信

当前同时兼容：

- Flutter WebView：`window.FlutterTrainingGame.postMessage`
- UniApp：`window.uni.postMessage`
- 微信小程序 web-view：`window.wx.miniProgram.postMessage`
- 普通浏览器/iframe：`window.parent.postMessage`
