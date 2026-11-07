import SwiftUI

/// Foundation LivingTabBar. A floating Liquid Glass bar for compact width: at regular width use the system's sidebar.
/// The lens stretches to a new tab: its leading edge moves first and the other edge follows on a spring. Dragging across
/// the bar moves the lens with the finger, resists past the ends and settles on the nearest tab. While content scrolls down
/// the bar shrinks to its icons: every tab stays, the lens stays on the selected one, the bar keeps its place.
/// An optional tint colors the lens, for example the accent of the content on screen.
/// The caller owns `isCollapsed` and `tint`, and hosts the bar above the content with `safeAreaInset(edge: .bottom)`,
/// so its bottom edge sits on the bottom safe-area inset. Requires iOS 26.
/// Usage: `FLivingTabBar(tabs: tabs, selection: $tab, isCollapsed: $collapsed, tint: accent)`
@available(iOS 26.0, *)
public struct FLivingTabBar: View {
    public struct Tab: Identifiable {
        public let id: Int
        public let title: String
        public let systemImage: String

        public init(id: Int, title: String, systemImage: String) {
            self.id = id
            self.title = title
            self.systemImage = systemImage
        }
    }

    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Environment(\.accessibilityReduceTransparency) private var reduceTransparency

    private let tabs: [Tab]
    @Binding private var selection: Int
    @Binding private var isCollapsed: Bool
    private let tint: Color?

    @State private var frames: [Int: CGRect] = [:]
    @State private var lensLeading: CGFloat = 0
    @State private var lensTrailing: CGFloat = 0
    @State private var isDragging = false

    private nonisolated static let space = "FLivingTabBar"

    public init(tabs: [Tab], selection: Binding<Int>, isCollapsed: Binding<Bool>, tint: Color? = nil) {
        self.tabs = tabs
        self._selection = selection
        self._isCollapsed = isCollapsed
        self.tint = tint
    }

    public var body: some View {
        GlassEffectContainer {
            material(
                HStack(spacing: FoundationTokens.Space.s1) {
                    ForEach(tabs) { tab in
                        button(for: tab)
                    }
                }
                .padding(FoundationTokens.Space.s1)
                .background(alignment: .topLeading) { lens }
                .coordinateSpace(.named(Self.space))
            )
            .scaleEffect(isDragging && !reduceMotion ? 1.035 : 1)
            .gesture(drag)
        }
        .animation(motion, value: isCollapsed)
        .animation(motion, value: isDragging)
        .sensoryFeedback(.selection, trigger: selection)
        .onChange(of: selection) { old, new in
            moveLens(from: old, to: new)
        }
        .accessibilityElement(children: .contain)
        .accessibilityAddTraits(.isTabBar)
    }

    // Glass, or a solid ground when the user asked for less transparency.
    @ViewBuilder
    private func material(_ content: some View) -> some View {
        if reduceTransparency {
            content.background(FoundationTokens.Color.surface2, in: .capsule)
        } else {
            content.glassEffect(.regular.interactive(), in: .capsule)
        }
    }

    private var lens: some View {
        let height = frames[selection]?.height ?? 0
        let top = frames[selection]?.minY ?? 0
        return Capsule()
            .fill(lensColor)
            .overlay(Capsule().strokeBorder(.white.opacity(0.3), lineWidth: 0.5))
            .frame(width: max(lensTrailing - lensLeading, 0), height: height)
            .offset(x: lensLeading, y: top)
            .animation(.easeInOut(duration: FoundationTokens.Motion.base * 1.6), value: tint)
            .accessibilityHidden(true)
    }

    private var lensColor: Color {
        if let tint {
            return tint.opacity(0.26)
        }
        return FoundationTokens.Color.text1.opacity(0.14)
    }

    private var motion: Animation? {
        reduceMotion ? nil : .spring(duration: FoundationTokens.Motion.base * 1.5, bounce: 0.2)
    }

    private func button(for tab: Tab) -> some View {
        let isSelected = tab.id == selection
        return Button {
            if isCollapsed && isSelected {
                isCollapsed = false
            } else {
                selection = tab.id
            }
        } label: {
            VStack(spacing: FoundationTokens.Space.s1 / 2) {
                Image(systemName: tab.systemImage)
                    .symbolVariant(isSelected ? .fill : .none)
                    .contentTransition(.symbolEffect(.replace))
                    .imageScale(.large)
                if !isCollapsed {
                    Text(tab.title)
                        .font(FoundationTokens.Typography.caption)
                        .dynamicTypeSize(...DynamicTypeSize.xxxLarge)
                        .lineLimit(1)
                }
            }
            .foregroundStyle(isSelected ? FoundationTokens.Color.brandStrong : FoundationTokens.Color.text1)
            .frame(minWidth: FoundationTokens.Space.s6 * 1.5, minHeight: FoundationTokens.Space.s6 * 1.5)
            .padding(.horizontal, isCollapsed ? 0 : FoundationTokens.Space.s6)
            .contentShape(Capsule())
        }
        .buttonStyle(.plain)
        .onGeometryChange(for: CGRect.self) { proxy in
            proxy.frame(in: .named(Self.space))
        } action: { frame in
            frames[tab.id] = frame
            if tab.id == selection && !isDragging {
                placeLens(on: frame)
            }
        }
        .accessibilityLabel(tab.title)
        .accessibilityAddTraits(isSelected ? .isSelected : [])
        .accessibilityShowsLargeContentViewer {
            Label(tab.title, systemImage: tab.systemImage)
        }
    }

    // Layout changes (first layout, shrinking, growing) move both edges together.
    private func placeLens(on frame: CGRect) {
        withAnimation(motion) {
            lensLeading = frame.minX
            lensTrailing = frame.maxX
        }
    }

    // A new tab: the edge on the side of travel leads, the other follows on a softer spring, so the lens stretches.
    private func moveLens(from old: Int, to new: Int) {
        guard let target = frames[new] else { return }
        guard !reduceMotion else {
            lensLeading = target.minX
            lensTrailing = target.maxX
            return
        }
        let lead = Animation.spring(duration: FoundationTokens.Motion.base * 1.1, bounce: 0)
        let follow = Animation.spring(duration: FoundationTokens.Motion.base * 1.8, bounce: 0.2)
        let rightward = target.midX > (lensLeading + lensTrailing) / 2
        withAnimation(rightward ? lead : follow) { lensTrailing = target.maxX }
        withAnimation(rightward ? follow : lead) { lensLeading = target.minX }
    }

    // The lens follows the finger 1:1, resists past the ends, and settles where the gesture was heading.
    private var drag: some Gesture {
        DragGesture(minimumDistance: 10, coordinateSpace: .named(Self.space))
            .onChanged { value in
                guard let first = tabs.first.flatMap({ frames[$0.id] }),
                      let last = tabs.last.flatMap({ frames[$0.id] }) else { return }
                isDragging = true
                let width = lensTrailing - lensLeading
                var leading = value.location.x - width / 2
                if leading < first.minX {
                    leading = first.minX - rubberBand(first.minX - leading, width)
                } else if leading + width > last.maxX {
                    leading = last.maxX - width + rubberBand(leading + width - last.maxX, width)
                }
                lensLeading = leading
                lensTrailing = leading + width
            }
            .onEnded { value in
                isDragging = false
                let heading = value.predictedEndLocation.x
                let nearest = tabs.min { a, b in
                    abs((frames[a.id]?.midX ?? 0) - heading) < abs((frames[b.id]?.midX ?? 0) - heading)
                }
                if let nearest, nearest.id != selection {
                    selection = nearest.id
                } else if let frame = frames[selection] {
                    withAnimation(reduceMotion ? nil : .spring(duration: FoundationTokens.Motion.base * 1.8, bounce: 0.2)) {
                        lensLeading = frame.minX
                        lensTrailing = frame.maxX
                    }
                }
            }
    }

    private func rubberBand(_ overshoot: CGFloat, _ dimension: CGFloat) -> CGFloat {
        let constant: CGFloat = 0.55
        return (overshoot * dimension * constant) / (dimension + constant * abs(overshoot))
    }
}

@available(iOS 26.0, *)
#Preview {
    @Previewable @State var selection = 0
    @Previewable @State var isCollapsed = false
    VStack(spacing: 24) {
        FLivingTabBar(
            tabs: [
                .init(id: 0, title: "Trends", systemImage: "flame"),
                .init(id: 1, title: "My results", systemImage: "photo.on.rectangle")
            ],
            selection: $selection,
            isCollapsed: $isCollapsed,
            tint: selection == 0 ? .orange : .cyan
        )
        Toggle("Collapsed", isOn: $isCollapsed).padding(.horizontal, 48)
    }
}
