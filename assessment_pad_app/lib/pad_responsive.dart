import 'dart:math' as math;

import 'package:flutter/material.dart';

const double padDesignHeight = 768;
const double padMinDesignWidth = 1024;

double padDesignWidthForViewport(double width, double height) {
  final double safeHeight = height > 0 ? height : padDesignHeight;
  final double safeWidth = width > 0 ? width : padMinDesignWidth;
  return math.max(padMinDesignWidth, padDesignHeight * safeWidth / safeHeight);
}

class PadDialogViewport extends StatelessWidget {
  const PadDialogViewport({
    required this.child,
    this.alignment = Alignment.center,
    super.key,
  });

  final Widget child;
  final Alignment alignment;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final Size screenSize = MediaQuery.sizeOf(context);
        final double viewportWidth = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : screenSize.width;
        final double viewportHeight = constraints.maxHeight.isFinite
            ? constraints.maxHeight
            : screenSize.height;
        final double designWidth = padDesignWidthForViewport(
          viewportWidth,
          viewportHeight,
        );
        final MediaQueryData mediaQuery = MediaQuery.of(context).copyWith(
          size: Size(designWidth, padDesignHeight),
        );

        return Align(
          alignment: alignment,
          child: FittedBox(
            fit: BoxFit.contain,
            alignment: alignment,
            child: MediaQuery(
              data: mediaQuery,
              child: SizedBox(
                width: designWidth,
                height: padDesignHeight,
                child: child,
              ),
            ),
          ),
        );
      },
    );
  }
}
