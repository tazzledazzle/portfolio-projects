export class EventRepository {
    buffer = [];
    insert(event) {
        this.buffer.push(event);
    }
    list() {
        return this.buffer;
    }
}
