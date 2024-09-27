import {TodoService, TodoServiceListRequest} from "./gen";

import { createPromiseClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';

export const baseUrl = `http://127.0.0.1:3000`;

const transport = createConnectTransport({
    baseUrl,
    httpVersion: '1.1',
    interceptors: [
        (next) => {
            return async (req) => {
                console.log(`${req.url} [${req.service.typeName}]`);
                return next(req);
            };
        },
    ],
});

const api = {
    todos: createPromiseClient(TodoService, transport),
};

(async () => {
    const result = await api.todos.list(new TodoServiceListRequest({}));
    console.dir({result}, {depth: null});
})().catch(console.error);