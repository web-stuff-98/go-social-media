import Layout from "./components/layout/Layout";
import { AuthProvider } from "./context/AuthContext";
import { FileSocketProvider } from "./context/AttachmentContext";
import { InterfaceProvider } from "./context/InterfaceContext";
import { ModalProvider } from "./context/ModalContext";
import { MouseProvider } from "./context/MouseContext";
import { PostsProvider } from "./context/PostsContext";
import { SocketProvider } from "./context/SocketContext";
import { UserdropdownProvider } from "./context/UserdropdownContext";
import { UsersProvider } from "./context/UsersContext";

function App() {
  return (
    <InterfaceProvider>
      <MouseProvider>
        <SocketProvider>
          <UserdropdownProvider>
            <AuthProvider>
              <FileSocketProvider>
                <ModalProvider>
                  <PostsProvider>
                    <UsersProvider>
                      <Layout />
                    </UsersProvider>
                  </PostsProvider>
                </ModalProvider>
              </FileSocketProvider>
            </AuthProvider>
          </UserdropdownProvider>
        </SocketProvider>
      </MouseProvider>
    </InterfaceProvider>
  );
}

export default App;
