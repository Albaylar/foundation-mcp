import SwiftUI

/// Foundation CompareSlider. Two pictures of the same size, before and after, compared by dragging a divider across them.
/// The after picture shows to the right of the divider. The divider follows the finger and stays where it is let go;
/// `lineTop` starts its line, and the Before and After tags beside it, below a bar over the top of the picture;
/// `lineReach` stops the line short of the bottom when words sit over the lower part of the picture.
/// VoiceOver reads it as one adjustable element: swiping up or down moves the divider by a quarter. Requires iOS 26.
/// The caller decides which gesture wins near the screen's left edge (the system's back swipe).
/// Usage: `FCompareSlider(position: $pos, label: "Before and after", beforeTitle: "Before", afterTitle: "After", lineTop: 108) { before } after: { after }`
@available(iOS 26.0, *)
public struct FCompareSlider<Before: View, After: View>: View {
    @Environment(\.accessibilityReduceTransparency) private var reduceTransparency

    @Binding private var position: Double
    private let label: String
    private let beforeTitle: String
    private let afterTitle: String
    private let lineTop: CGFloat
    private let lineReach: CGFloat
    private let before: Before
    private let after: After

    @State private var isDragging = false

    /// `position` runs from 0 (all after) to 1 (all before); 0.5 is the middle.
    public init(
        position: Binding<Double>,
        label: String,
        beforeTitle: String,
        afterTitle: String,
        lineTop: CGFloat = 0,
        lineReach: CGFloat = 1,
        @ViewBuilder before: () -> Before,
        @ViewBuilder after: () -> After
    ) {
        self._position = position
        self.label = label
        self.beforeTitle = beforeTitle
        self.afterTitle = afterTitle
        self.lineTop = max(lineTop, 0)
        self.lineReach = min(max(lineReach, 0), 1)
        self.before = before()
        self.after = after()
    }

    public var body: some View {
        GeometryReader { proxy in
            let width = proxy.size.width
            let x = width * position
            let lineHeight = max(proxy.size.height * lineReach - lineTop, 0)
            ZStack(alignment: .topLeading) {
                before
                    .frame(width: width, height: proxy.size.height)
                after
                    .frame(width: width, height: proxy.size.height)
                    .mask(alignment: .trailing) {
                        Rectangle().frame(width: max(width - x, 0))
                    }
                Rectangle()
                    .fill(FoundationTokens.Color.onMedia)
                    .frame(width: 2, height: lineHeight)
                    .shadow(color: FoundationTokens.Color.mediaGround.opacity(0.35), radius: 6)
                    .offset(x: x - 1, y: lineTop)
                handle
                    .offset(x: x - handleSize / 2, y: lineTop + lineHeight * 0.56 - handleSize / 2)
                tags
            }
            .clipped()
            .contentShape(Rectangle())
            .gesture(
                DragGesture(minimumDistance: 0)
                    .onChanged { value in
                        isDragging = true
                        position = min(max(value.location.x / max(width, 1), 0), 1)
                    }
                    .onEnded { _ in isDragging = false }
            )
        }
        .sensoryFeedback(.selection, trigger: isDragging)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(label)
        .accessibilityValue(Text(position, format: .percent.precision(.fractionLength(0))))
        .accessibilityAdjustableAction { direction in
            switch direction {
            case .increment: position = min(position + 0.25, 1)
            case .decrement: position = max(position - 0.25, 0)
            @unknown default: break
            }
        }
    }

    private var handleSize: CGFloat { FoundationTokens.Space.s6 + FoundationTokens.Space.s5 - FoundationTokens.Space.s1 }

    private var handle: some View {
        Image(systemName: "arrow.left.and.right")
            .font(FoundationTokens.Typography.label)
            .foregroundStyle(FoundationTokens.Color.onMedia)
            .frame(width: handleSize, height: handleSize)
            .background { handleGround }
            .scaleEffect(isDragging ? 1.08 : 1)
            .animation(.spring(duration: FoundationTokens.Motion.fast * 2), value: isDragging)
    }

    @ViewBuilder
    private var handleGround: some View {
        if reduceTransparency {
            Circle().fill(FoundationTokens.Color.mediaGround.opacity(0.9))
        } else {
            Circle().fill(FoundationTokens.Color.mediaGround.opacity(0.2)).glassEffect(.regular.interactive(), in: .circle)
        }
    }

    // "Before" and "After" on a dark ground, so they read over any picture.
    private var tags: some View {
        HStack {
            tag(beforeTitle)
            Spacer(minLength: FoundationTokens.Space.s2)
            tag(afterTitle)
        }
        .padding(.horizontal, FoundationTokens.Space.s4)
        // Under a bar over the picture (the status bar, a counter) the tags start where the line starts.
        .padding(.top, max(FoundationTokens.Space.s4, lineTop + FoundationTokens.Space.s2))
        .accessibilityHidden(true)
    }

    private func tag(_ text: String) -> some View {
        Text(text)
            .font(FoundationTokens.Typography.caption.weight(.semibold))
            .foregroundStyle(FoundationTokens.Color.onMedia)
            .padding(.vertical, FoundationTokens.Space.s1)
            .padding(.horizontal, FoundationTokens.Space.s3)
            .background(FoundationTokens.Color.mediaGround.opacity(0.8), in: Capsule())
            .dynamicTypeSize(...DynamicTypeSize.accessibility2)
    }
}

@available(iOS 26.0, *)
#Preview {
    @Previewable @State var position = 0.5
    FCompareSlider(position: $position, label: "Before and after", beforeTitle: "Before", afterTitle: "After") {
        LinearGradient(colors: [.gray, .black], startPoint: .top, endPoint: .bottom)
    } after: {
        LinearGradient(colors: [.orange, .purple], startPoint: .top, endPoint: .bottom)
    }
    .frame(height: 420)
}
