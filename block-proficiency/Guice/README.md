- Writes `AbstractModule`; uses `@Inject` (constructor, field, method injection)

- Understands bindings: `bind(Interface.class).to(Impl.class)`, `toInstance`, `toProvider`

- Scopes: `@Singleton`, `@RequestScoped`; understands lifecycle implications

- Uses `@Named` and `@Qualifier` annotations for multi-binding disambiguation
