package rules

import io.gitlab.arturbosch.detekt.api.CodeSmell
import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.api.Debt
import io.gitlab.arturbosch.detekt.api.Entity
import io.gitlab.arturbosch.detekt.api.Issue
import io.gitlab.arturbosch.detekt.api.Rule
import io.gitlab.arturbosch.detekt.api.Severity
import org.jetbrains.kotlin.psi.KtNamedFunction

class MissingTransactionalAnnotationRule(config: Config) : Rule(config) {
    override val issue: Issue =
        Issue("MissingTransactionalAnnotation", Severity.Defect, "Mutating service methods should be @Transactional.", Debt.TEN_MINS)

    override fun visitNamedFunction(function: KtNamedFunction) {
        super.visitNamedFunction(function)
        val name = function.name ?: return
        val mutating = name.startsWith("create") || name.startsWith("update") || name.startsWith("delete")
        val annotated = function.annotationEntries.any { it.shortName?.asString() == "Transactional" }
        if (mutating && !annotated) {
            report(CodeSmell(issue, Entity.from(function), "Add @Transactional to '$name'."))
        }
    }
}
