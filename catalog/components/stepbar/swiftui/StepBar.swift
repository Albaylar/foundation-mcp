import SwiftUI

/// Foundation StepBar. The top bar of a short run of steps, such as a funnel: back, where the user is, and close,
/// together on one Liquid Glass island. Up to eight steps the middle shows one segment per step, the steps behind and
/// the current one filled with the tint; more than eight, one track and a count, so the position stays readable.
/// VoiceOver reads back, the progress (the label the caller gives, "Step 2 of 4") and close, in that order.
/// Requires iOS 26. Usage:
/// `FStepBar(current: 1, total: 4, tint: accent, label: "Step 2 of 4", backLabel: "Back", closeLabel: "Close", onBack: back, onClose: close)`
@available(iOS 26.0, *)
public struct FStepBar: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Environment(\.accessibilityReduceTransparency) private var reduceTransparency

    private let current: Int
    private let total: Int
    private let tint: Color
    private let label: String
    private let backLabel: String
    private let closeLabel: String
    private let onBack: () -> Void
    private let onClose: () -> Void

    /// `current` counts from 0. `tint` fills the progress, for example an item's accent color.
    public init(
        current: Int,
        total: Int,
        tint: Color = FoundationTokens.Color.brand,
        label: String,
        backLabel: String,
        closeLabel: String,
        onBack: @escaping () -> Void,
        onClose: @escaping () -> Void
    ) {
        self.total = max(total, 1)
        self.current = min(max(current, 0), max(total, 1) - 1)
        self.tint = tint
        self.label = label
        self.backLabel = backLabel
        self.closeLabel = closeLabel
        self.onBack = onBack
        self.onClose = onClose
    }

    public var body: some View {
        HStack(spacing: FoundationTokens.Space.s2) {
            button("chevron.left", backLabel, onBack)
            progress
                .frame(maxWidth: .infinity)
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(label)
            button("xmark", closeLabel, onClose)
        }
        .padding(FoundationTokens.Space.s1)
        .background { ground }
        .animation(reduceMotion ? nil : .easeOut(duration: FoundationTokens.Motion.base * 2), value: current)
        .accessibilityElement(children: .contain)
    }

    @ViewBuilder
    private var progress: some View {
        if total <= 8 {
            HStack(spacing: FoundationTokens.Space.s1) {
                ForEach(0..<total, id: \.self) { step in
                    Capsule()
                        .fill(step <= current ? tint : FoundationTokens.Color.text1.opacity(0.2))
                        .frame(height: FoundationTokens.Space.s1)
                }
            }
        } else {
            HStack(spacing: FoundationTokens.Space.s3) {
                GeometryReader { proxy in
                    ZStack(alignment: .leading) {
                        Capsule().fill(FoundationTokens.Color.text1.opacity(0.2))
                        Capsule().fill(tint)
                            .frame(width: proxy.size.width * CGFloat(current + 1) / CGFloat(total))
                    }
                }
                .frame(height: FoundationTokens.Space.s1)
                Text("\(current + 1) / \(total)")
                    .font(FoundationTokens.Typography.caption.weight(.semibold))
                    .monospacedDigit()
                    .foregroundStyle(FoundationTokens.Color.text1)
                    .dynamicTypeSize(...DynamicTypeSize.xxxLarge)
            }
        }
    }

    private func button(_ symbol: String, _ spoken: String, _ action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Image(systemName: symbol)
                .font(FoundationTokens.Typography.label)
                .imageScale(.large)
                .frame(width: FoundationTokens.Space.s6 + FoundationTokens.Space.s3, height: FoundationTokens.Space.s6 + FoundationTokens.Space.s3)
                .contentShape(Circle())
        }
        .buttonStyle(PressStyle())
        .foregroundStyle(FoundationTokens.Color.text1)
        .accessibilityLabel(spoken)
    }

    // Glass, or a solid ground when the user asked for less transparency.
    @ViewBuilder
    private var ground: some View {
        if reduceTransparency {
            Capsule().fill(FoundationTokens.Color.surface2)
        } else {
            Capsule().fill(.clear).glassEffect(.regular, in: .capsule)
        }
    }

    private struct PressStyle: ButtonStyle {
        func makeBody(configuration: Configuration) -> some View {
            configuration.label
                .background(Circle().fill(FoundationTokens.Color.text1.opacity(configuration.isPressed ? 0.12 : 0)))
                .scaleEffect(configuration.isPressed ? 0.94 : 1)
                .animation(.easeOut(duration: FoundationTokens.Motion.fast), value: configuration.isPressed)
        }
    }
}

@available(iOS 26.0, *)
#Preview {
    VStack(spacing: FoundationTokens.Space.s4) {
        FStepBar(current: 1, total: 4, label: "Step 2 of 4", backLabel: "Back", closeLabel: "Close", onBack: {}, onClose: {})
        FStepBar(current: 3, total: 12, label: "Step 4 of 12", backLabel: "Back", closeLabel: "Close", onBack: {}, onClose: {})
    }
    .padding(FoundationTokens.Space.s4)
    .background(FoundationTokens.Color.mediaGround)
}
