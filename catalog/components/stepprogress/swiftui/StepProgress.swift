import SwiftUI

/// Foundation StepProgress. Where the user is in a short run of steps, on a glass capsule in the navigation layer.
/// Up to eight steps it shows one segment per step: the steps behind and the current one filled with the tint.
/// More than eight: one track filled to the position and a count, so the position stays readable.
/// Not a control: VoiceOver reads it as one element with the label the caller gives ("Step 2 of 4"). Requires iOS 26.
/// Usage: `FStepProgress(current: 1, total: 4, tint: accent, label: "Step 2 of 4")`
@available(iOS 26.0, *)
public struct FStepProgress: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Environment(\.accessibilityReduceTransparency) private var reduceTransparency

    private let current: Int
    private let total: Int
    private let tint: Color
    private let label: String

    /// `current` counts from 0. `tint` must stand out against the dark glass, for example an item's accent color.
    public init(current: Int, total: Int, tint: Color = FoundationTokens.Color.brand, label: String) {
        self.total = max(total, 1)
        self.current = min(max(current, 0), max(total, 1) - 1)
        self.tint = tint
        self.label = label
    }

    public var body: some View {
        content
            .padding(.horizontal, FoundationTokens.Space.s4)
            .frame(minHeight: FoundationTokens.Space.s6 + FoundationTokens.Space.s3)
            .background { ground }
            .animation(reduceMotion ? nil : .easeOut(duration: FoundationTokens.Motion.base * 2), value: current)
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(label)
    }

    @ViewBuilder
    private var content: some View {
        if total <= 8 {
            HStack(spacing: FoundationTokens.Space.s1) {
                ForEach(0..<total, id: \.self) { step in
                    Capsule()
                        .fill(step <= current ? tint : FoundationTokens.Color.onMedia.opacity(0.26))
                        .frame(height: FoundationTokens.Space.s1)
                }
            }
        } else {
            HStack(spacing: FoundationTokens.Space.s3) {
                GeometryReader { proxy in
                    ZStack(alignment: .leading) {
                        Capsule().fill(FoundationTokens.Color.onMedia.opacity(0.26))
                        Capsule().fill(tint)
                            .frame(width: proxy.size.width * CGFloat(current + 1) / CGFloat(total))
                    }
                }
                .frame(height: FoundationTokens.Space.s1)
                Text("\(current + 1) / \(total)")
                    .font(FoundationTokens.Typography.caption.weight(.semibold))
                    .monospacedDigit()
                    .foregroundStyle(FoundationTokens.Color.onMedia)
                    .dynamicTypeSize(...DynamicTypeSize.xxxLarge)
            }
        }
    }

    // Glass, or a solid dark ground when the user asked for less transparency. Dark in both appearances, like media.
    @ViewBuilder
    private var ground: some View {
        if reduceTransparency {
            Capsule().fill(FoundationTokens.Color.mediaGround.opacity(0.9))
        } else {
            Capsule().fill(FoundationTokens.Color.mediaGround.opacity(0.3))
                .glassEffect(.regular, in: .capsule)
        }
    }
}

@available(iOS 26.0, *)
#Preview {
    VStack(spacing: FoundationTokens.Space.s4) {
        FStepProgress(current: 1, total: 4, label: "Step 2 of 4")
        FStepProgress(current: 3, total: 12, label: "Step 4 of 12")
    }
    .padding(FoundationTokens.Space.s4)
    .background(FoundationTokens.Color.mediaGround)
}
