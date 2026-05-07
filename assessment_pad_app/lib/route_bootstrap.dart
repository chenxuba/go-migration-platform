import 'dart:async';

import 'package:flutter/material.dart';

void runAfterRouteEntrance(
  BuildContext context,
  FutureOr<void> Function() action,
) {
  WidgetsBinding.instance.addPostFrameCallback((_) {
    if (!context.mounted) {
      return;
    }
    final ModalRoute<dynamic>? route = ModalRoute.of(context);
    final Animation<double>? animation = route?.animation;
    if (animation == null ||
        animation.isCompleted ||
        animation.status == AnimationStatus.completed) {
      unawaited(Future<void>.sync(action));
      return;
    }
    late final AnimationStatusListener listener;
    listener = (AnimationStatus status) {
      if (status != AnimationStatus.completed) {
        return;
      }
      animation.removeStatusListener(listener);
      if (!context.mounted) {
        return;
      }
      unawaited(Future<void>.sync(action));
    };
    animation.addStatusListener(listener);
  });
}
