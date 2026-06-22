package rules

import io.gitlab.arturbosch.detekt.api.CodeSmell
import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.api.Debt
import io.gitlab.arturbosch.detekt.api.Entity
import io.gitlab.arturbosch.detekt.api.Issue
import io.gitlab.arturbosch.detekt.api.Rule
import io.gitlab.arturbosch.detekt.api.Severity
import org.jetbrains.kotlin.psi.KtBinaryExpressionWithTypeRHS

class UncheckedPlatformTypeCastRule(config: Config) : Rule(config) {
    override val issue: Issue =
        Issue("UncheckedPlatformTypeCast", Severity.Warning, "Avoid unchecked casts around platform types.", Debt.FIVE_MINS)

    override fun visitBinaryWithTypeRHSExpression(expression: KtBinaryExpressionWithTypeRHS) {
        super.visitBinaryWithTypeRHSExpression(expression)
        if (expression.operationReference.text == "as") {
            report(CodeSmell(issue, Entity.from(expression), "Use safe cast or null guards before casting platform type."))
        }
    }
}
