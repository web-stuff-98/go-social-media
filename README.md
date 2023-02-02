# https://go-social-media-js.herokuapp.com

This is the second version of my Go project, and it actually works. Its also the successor to Prisma-Social-Media which was quite messy and probably much slower than this.

I already completed the project doing manual testing, I just wrote basic unit tests that check if things are rendering and requests are being made because I keep getting errors that don't have any solutions online when I try to write unit
tests for anything complex. I will improve the tests if I can figure out how to resolve annoying Jest errors.

I also added a file called DOC.md which is supposed to explain how everything works, mainly for how the server and client send, receive messages and how the client infer types from inbound messages.

I may also replace the React Context API with Redux and maybe replace my useEffect API calls with SWR or React Query

It uses net/http, gorilla mux and gorilla websocket, you can video chat in groups or in private and download and upload attachments and watch the progress bar. It uploads the files in chunks, uploads are also really slow because it's using MongoDB for storage on the free tier.

### Stuff

- Group & private video chat using simple-peer
- Nested comments
- Types & models for socket events
- Context API & Reducers
- Formik
- File attachments
- Rate limiting
- Pagination/Filtering/Sorting
- Updates & DB cleanup via changestreams
- Recursion for deleting orphans & downloading files
- Refresh tokens
- Live updates for everything
