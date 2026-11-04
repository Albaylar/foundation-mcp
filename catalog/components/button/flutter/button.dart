import 'package:flutter/material.dart';

import '../tokens/foundation_tokens.dart';

enum FButtonVariant { primary, secondary, destructive }

/// Foundation Button. Colors and spacing come from [FoundationTokens] (generated from tokens.json).
/// Usage: `FButton(title: 'Save', isLoading: saving, onPressed: save)`
class FButton extends StatelessWidget {
  const FButton({
    super.key,
    required this.title,
    required this.onPressed,
    this.variant = FButtonVariant.primary,
    this.isLoading = false,
  });

  final String title;
  final VoidCallback? onPressed;
  final FButtonVariant variant;
  final bool isLoading;

  Color get _background => switch (variant) {
        FButtonVariant.primary => FoundationTokens.colorBrand,
        FButtonVariant.secondary => FoundationTokens.colorSurface2,
        FButtonVariant.destructive => FoundationTokens.colorDanger,
      };

  Color get _foreground => switch (variant) {
        FButtonVariant.primary => FoundationTokens.colorOnBrand,
        FButtonVariant.secondary => FoundationTokens.colorText1,
        FButtonVariant.destructive => FoundationTokens.colorOnDanger,
      };

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: title,
      value: isLoading ? 'loading' : null,
      child: SizedBox(
        height: 48,
        width: double.infinity,
        child: FilledButton(
          onPressed: isLoading ? null : onPressed,
          style: FilledButton.styleFrom(
            backgroundColor: _background,
            foregroundColor: _foreground,
            disabledBackgroundColor: _background,
            padding: const EdgeInsets.symmetric(horizontal: FoundationTokens.space4),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(FoundationTokens.radiusField)),
            textStyle: FoundationTokens.typeLabel,
          ),
          child: AnimatedSwitcher(
            duration: FoundationTokens.motionFast,
            child: isLoading
                ? SizedBox.square(dimension: 20, child: CircularProgressIndicator(strokeWidth: 2, color: _foreground))
                : Text(title),
          ),
        ),
      ),
    );
  }
}
