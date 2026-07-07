## 2. Kotlin

- Understands Kotlin's null safety (`?`, `!!`, `?.`, `?:`)

nullability benefits, use cases

- Can read data classes, object declarations, companion objects
```kotlin

data class Clazz(val attribute: Bool)

object Fun() {

}
class Father {
    init {}
    companion object() {}
}

```
- Knows Kotlin compiles to JVM and is interoperable with Java



### Level 2 — Practitioner
- Idiomatic use of: extension functions, sealed classes, `when` expressions, destructuring

- Coroutines: launches with `launch`/`async`, understands `suspend`, uses `withContext`

- Uses `Flow` for reactive streams; handles `StateFlow`/`SharedFlow`

- Scope functions: `let`, `run`, `apply`, `also`, `with` — uses each appropriately

- Kotlin collections API: `map`, `filter`, `fold`, `groupBy`, `associate`, etc.

- Writes clean DSLs using lambda receivers and infix functions

### Level 3 — Proficient
- Structured concurrency: `CoroutineScope`, `SupervisorJob`, `CoroutineExceptionHandler`
- Cancellation propagation, cooperative cancellation with `isActive`/`ensureActive`
- Understands coroutine internals: continuation-passing style (CPS), state machines
- Designs type-safe builders and DSLs
- Uses `inline`/`reified` generics correctly for type erasure workarounds
- Applies `@JvmOverloads`, `@JvmStatic`, `@JvmField` for clean Java interop
- Delegates: `by lazy`, `by Delegates.observable`, property delegation contract

### Level 4 — Expert
- Compiler plugin authorship (K2 plugins, IR transforms)
- Deep coroutine dispatcher internals; writes custom dispatchers
- Kotlin Multiplatform (KMP): shared business logic, `expect`/`actual`, sourceset configuration
- Understands desugaring and optimization differences vs. Java equivalents
- Uses KSP/KAPT for annotation processing and code generation