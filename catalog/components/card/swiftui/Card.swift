import SwiftUI

/// Foundation Card. Usage: `FCard { Text("…") }` or `FCard(action: open) { … }`
public struct FCard<Content: View>: View {
    private let action: (() -> Void)?
    private let content: Content

    public init(action: (() -> Void)? = nil, @ViewBuilder content: () -> Content) {
        self.action = action
        self.content = content()
    }

    public var body: some View {
        if let action {
            Button(action: action, label: { surface }).buttonStyle(.plain).accessibilityElement(children: .combine)
        } else {
            surface
        }
    }

    private var surface: some View {
        content
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(FoundationTokens.Space.s4)
            .background(FoundationTokens.Color.surface, in: RoundedRectangle(cornerRadius: FoundationTokens.Radius.card))
            .overlay(RoundedRectangle(cornerRadius: FoundationTokens.Radius.card).stroke(FoundationTokens.Color.line, lineWidth: 1))
    }
}
