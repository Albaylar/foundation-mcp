import SwiftUI

/// Foundation LivingTabBar. A floating Liquid Glass bar: the lens follows the selected tab on a spring, the bar shrinks to the
/// selected tab while content scrolls down. For compact width only: at regular width use the system's sidebar.
/// The caller owns `isCollapsed`, and hosts it above the content with `safeAreaInset(edge: .bottom)`. Requires iOS 26.
/// Usage: `FLivingTabBar(tabs: tabs, selection: $tab, isCollapsed: $collapsed)`
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
    @Namespace private var lens

    private let tabs: [Tab]
    @Binding private var selection: Int
    @Binding private var isCollapsed: Bool

    public init(tabs: [Tab], selection: Binding<Int>, isCollapsed: Binding<Bool>) {
        self.tabs = tabs
        self._selection = selection
        self._isCollapsed = isCollapsed
    }

    public var body: some View {
        GlassEffectContainer {
            HStack(spacing: FoundationTokens.Space.s1) {
                ForEach(visibleTabs) { tab in
                    button(for: tab)
                }
            }
            .padding(FoundationTokens.Space.s1)
            .glassEffect(.regular.interactive(), in: .capsule)
        }
        .animation(motion, value: selection)
        .animation(motion, value: isCollapsed)
        .sensoryFeedback(.selection, trigger: selection)
    }

    private var visibleTabs: [Tab] {
        isCollapsed ? tabs.filter { $0.id == selection } : tabs
    }

    private var motion: Animation? {
        reduceMotion ? nil : .spring(duration: FoundationTokens.Motion.base * 1.5, bounce: 0.2)
    }

    private func button(for tab: Tab) -> some View {
        let isSelected = tab.id == selection
        return Button {
            if isCollapsed {
                isCollapsed = false
            } else {
                selection = tab.id
            }
        } label: {
            VStack(spacing: FoundationTokens.Space.s1 / 2) {
                Image(systemName: tab.systemImage)
                    .symbolVariant(isSelected ? .fill : .none)
                    .imageScale(.large)
                if !isCollapsed {
                    Text(tab.title).font(FoundationTokens.Typography.caption)
                }
            }
            .foregroundStyle(isSelected ? FoundationTokens.Color.brandStrong : FoundationTokens.Color.text1)
            .frame(minWidth: FoundationTokens.Space.s6 * 1.5, minHeight: FoundationTokens.Space.s6 * 1.5)
            .padding(.horizontal, isCollapsed ? 0 : FoundationTokens.Space.s6)
            .background {
                if isSelected {
                    Capsule()
                        .fill(FoundationTokens.Color.text1.opacity(0.14))
                        .matchedGeometryEffect(id: "lens", in: lens)
                }
            }
            .contentShape(Capsule())
        }
        .buttonStyle(.plain)
        .accessibilityLabel(tab.title)
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}

#Preview {
    @Previewable @State var selection = 0
    @Previewable @State var isCollapsed = false
    FLivingTabBar(
        tabs: [
            .init(id: 0, title: "Trends", systemImage: "flame"),
            .init(id: 1, title: "My results", systemImage: "photo.on.rectangle")
        ],
        selection: $selection,
        isCollapsed: $isCollapsed
    )
}
