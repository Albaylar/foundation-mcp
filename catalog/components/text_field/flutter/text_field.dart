import 'package:flutter/material.dart';

import '../tokens/foundation_tokens.dart';

/// Foundation TextField: label above, helper or error below.
/// Usage: `FTextField(label: 'E-Mail', controller: emailController, error: emailError)`
class FTextField extends StatelessWidget {
  const FTextField({super.key, required this.label, required this.controller, this.helper, this.error});

  final String label;
  final TextEditingController controller;
  final String? helper;
  final String? error;

  @override
  Widget build(BuildContext context) {
    final border = OutlineInputBorder(
      borderRadius: BorderRadius.circular(FoundationTokens.radiusField),
      borderSide: BorderSide(color: error == null ? FoundationTokens.colorLine : FoundationTokens.colorDanger),
    );
    return TextField(
      controller: controller,
      style: FoundationTokens.typeBody,
      decoration: InputDecoration(
        labelText: label,
        helperText: helper,
        errorText: error,
        helperStyle: FoundationTokens.typeCaption.copyWith(color: FoundationTokens.colorText2),
        errorStyle: FoundationTokens.typeCaption.copyWith(color: FoundationTokens.colorDanger),
        filled: true,
        fillColor: FoundationTokens.colorSurface,
        contentPadding: const EdgeInsets.all(FoundationTokens.space3),
        border: border,
        enabledBorder: border,
        focusedBorder: border.copyWith(borderSide: const BorderSide(color: FoundationTokens.colorBrand)),
      ),
    );
  }
}
