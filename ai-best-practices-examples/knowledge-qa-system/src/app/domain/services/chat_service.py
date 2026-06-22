from collections.abc import Iterator

from langchain_core.output_parsers import StrOutputParser
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.runnables import RunnableLambda

from app.domain.services.retrieval_service import RetrievalService
from app.schemas.chat import ChatResponse


class ChatService:
    def __init__(self, retrieval_service: RetrievalService | None = None) -> None:
        self.retrieval_service = retrieval_service or RetrievalService()

    def answer(self, message: str, top_k: int | None = None) -> ChatResponse:
        citations, snippets, partial_context = self.retrieval_service.retrieve(message, top_k=top_k)
        if partial_context:
            answer = "I do not have enough grounded context to answer confidently yet."
        else:
            context = "\n\n".join(snippets[:4])
            prompt = ChatPromptTemplate.from_template(
                "Question: {question}\n\nContext:\n{context}\n\n"
                "Answer using only the context in 2-4 concise sentences."
            )
            chain = (
                prompt
                | RunnableLambda(
                    lambda prompt_value: prompt_value.to_messages()[0].content  # grounded synthesis fallback
                )
                | RunnableLambda(lambda text: text.split("Context:\n", 1)[-1].strip())
                | RunnableLambda(lambda text: text.split("\n\nAnswer using only the context", 1)[0].strip())
                | RunnableLambda(lambda text: text[:420] if len(text) > 420 else text)
                | StrOutputParser()
            )
            answer = chain.invoke({"question": message, "context": context})
        return ChatResponse(answer=answer, citations=citations, partial_context=partial_context)

    def stream_answer(self, text: str) -> Iterator[str]:
        for token in text.split():
            yield f"{token} "
