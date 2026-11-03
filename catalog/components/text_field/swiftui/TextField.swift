import SwiftUI

/// Foundation TextField: label above, helper or error below.
/// Usage: `FTextField("E-Mail", text: $email, error: emailError)`
public struct FTextField: View {
    private let label: String
    @Binding private var text: String
    private let helper: String?
    private let error: String?

    public init(_ label: String, text: Binding<String>, helper: String? = nil, error: String? = nil) {
        self.label = label
        self._text = text
        self.helper = helper
        self.error = error
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: FoundationTokens.Space.s2) {
            Text(label).font(FoundationTokens.Typography.caption).foregroundStyle(FoundationTokens.Color.text2)
            TextField("", text: $text)
                .font(FoundationTokens.Typography.body)
                .padding(FoundationTokens.Space.s3)
                .frame(minHeight: 44)
                .background(FoundationTokens.Color.surface, in: RoundedRectangle(cornerRadius: FoundationTokens.Radius.field))
                .overlay(RoundedRectangle(cornerRadius: FoundationTokens.Radius.field).stroke(borderColor, lineWidth: 1))
            if let message = error ?? helper {
                Text(message).font(FoundationTokens.Typography.caption).foregroundStyle(error == nil ? FoundationTokens.Color.text2 : FoundationTokens.Color.danger)
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel(label)
        .accessibilityValue(error ?? text)
    }

    private var borderColor: Color { error == nil ? FoundationTokens.Color.line : FoundationTokens.Color.danger }
}
