import SwiftUI

/// Foundation Button. Colors and spacing come from `FoundationTokens` (generated from tokens.json).
/// Usage: `FButton("Save", variant: .primary, isLoading: saving) { save() }`
public struct FButton: View {
    public enum Variant { case primary, secondary, destructive }

    private let title: String
    private let variant: Variant
    private let isLoading: Bool
    private let action: () -> Void

    public init(_ title: String, variant: Variant = .primary, isLoading: Bool = false, action: @escaping () -> Void) {
        self.title = title
        self.variant = variant
        self.isLoading = isLoading
        self.action = action
    }

    public var body: some View {
        Button(action: action, label: {
            ZStack {
                Text(title)
                    .font(FoundationTokens.Typography.label)
                    .opacity(isLoading ? 0 : 1)
                if isLoading { ProgressView().tint(foreground) }
            }
            .frame(maxWidth: .infinity, minHeight: 44)
            .padding(.horizontal, FoundationTokens.Space.s4)
        })
        .foregroundStyle(foreground)
        .background(background, in: RoundedRectangle(cornerRadius: FoundationTokens.Radius.field))
        .disabled(isLoading)
        .accessibilityLabel(title)
        .accessibilityValue(isLoading ? Text("loading") : Text(""))
        .animation(.easeInOut(duration: FoundationTokens.Motion.fast), value: isLoading)
    }

    private var background: Color {
        switch variant {
        case .primary: FoundationTokens.Color.brand
        case .secondary: FoundationTokens.Color.surface2
        case .destructive: FoundationTokens.Color.danger
        }
    }

    private var foreground: Color {
        switch variant {
        case .primary: FoundationTokens.Color.onBrand
        case .secondary: FoundationTokens.Color.text1
        case .destructive: FoundationTokens.Color.onDanger
        }
    }
}

#Preview {
    VStack(spacing: FoundationTokens.Space.s3) {
        FButton("Save") {}
        FButton("Cancel", variant: .secondary) {}
        FButton("Delete", variant: .destructive) {}
        FButton("Saving", isLoading: true) {}
    }
    .padding()
}
