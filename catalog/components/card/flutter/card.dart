import 'package:flutter/material.dart';

import '../tokens/foundation_tokens.dart';

/// Foundation Card. Usage: `FCard(child: …)` or `FCard(onTap: open, child: …)`
class FCard extends StatelessWidget {
  const FCard({super.key, required this.child, this.onTap});

  final Widget child;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(FoundationTokens.radiusCard),
      side: const BorderSide(color: FoundationTokens.colorLine),
    );
    return Material(
      color: FoundationTokens.colorSurface,
      shape: shape,
      child: InkWell(
        onTap: onTap,
        customBorder: shape,
        child: Padding(padding: const EdgeInsets.all(FoundationTokens.space4), child: child),
      ),
    );
  }
}
