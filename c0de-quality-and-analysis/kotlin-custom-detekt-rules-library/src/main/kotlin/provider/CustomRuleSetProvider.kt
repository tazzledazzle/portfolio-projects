package provider

import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.api.RuleSet
import io.gitlab.arturbosch.detekt.api.RuleSetProvider
import rules.CoroutineScopeNamingRule
import rules.MissingTransactionalAnnotationRule
import rules.UncheckedPlatformTypeCastRule

class CustomRuleSetProvider : RuleSetProvider {
    override val ruleSetId: String = "team-custom-rules"

    override fun instance(config: Config): RuleSet =
        RuleSet(
            id = ruleSetId,
            rules = listOf(
                CoroutineScopeNamingRule(config),
                UncheckedPlatformTypeCastRule(config),
                MissingTransactionalAnnotationRule(config),
            ),
        )
}
