package rules

import io.gitlab.arturbosch.detekt.api.CodeSmell
import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.api.Debt
import io.gitlab.arturbosch.detekt.api.Entity
import io.gitlab.arturbosch.detekt.api.Issue
import io.gitlab.arturbosch.detekt.api.Rule
import io.gitlab.arturbosch.detekt.api.Severity
import org.jetbrains.kotlin.psi.KtProperty

class CoroutineScopeNamingRule(
    config: Config,
) : Rule(config) {
    override val issue: Issue =
        Issue("CoroutineScopeNaming", Severity.Style, "Coroutine scope properties should end with 'Scope'.", Debt.FIVE_MINS)

    override fun visitProperty(property: KtProperty) {
        super.visitProperty(property)
        val name = property.name ?: return
        if (name.contains("coroutine", ignoreCase = true) && !name.endsWith("Scope")) {
            report(CodeSmell(issue, Entity.from(property), "Rename '$name' to use Scope suffix."))
        }
    }
}
