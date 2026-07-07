# Kotlin Custom Detekt Rules Library

## Overview
This project provides team-specific `detekt` rules for Kotlin services and libraries. It extends static analysis with conventions not covered by default rulesets.

## Architecture
- `src/main/kotlin/rules`: custom `Rule` implementations.
- `src/main/kotlin/provider`: `RuleSetProvider` wiring for detekt runtime discovery.
- `src/test/kotlin`: rule tests using code fixtures.
- `config`: rule activation and threshold configuration.
- `.github/workflows`: CI checks and test execution.

## Use Cases
- Enforce coroutine scope naming consistency.
- Flag unchecked casts to Kotlin platform types.
- Detect missing `@Transactional` annotations on service-layer mutating methods.

## Usage
1. Build and publish the plugin jar.
2. Add the jar to consumer projects' detekt plugin classpath.
3. Enable rules in `config/detekt-custom-rules.yml`.
4. Run `./gradlew detekt`.

## Control Flow
1. Detekt loads `CustomRuleSetProvider`.
2. Provider registers all rule classes in `CustomRuleSet`.
3. Each rule visits Kotlin PSI elements and reports findings.
4. CI fails when findings exceed configured thresholds.

## Project Structure
```text
kotlin-custom-detekt-rules-library/
  .github/workflows/ci.yml
  config/detekt-custom-rules.yml
  scripts/run-local-detekt.sh
  src/main/kotlin/provider/CustomRuleSetProvider.kt
  src/main/kotlin/rules/CoroutineScopeNamingRule.kt
  src/main/kotlin/rules/MissingTransactionalAnnotationRule.kt
  src/main/kotlin/rules/UncheckedPlatformTypeCastRule.kt
  src/test/kotlin/rules/CustomRulesTest.kt
```
