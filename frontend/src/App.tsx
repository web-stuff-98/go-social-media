import Layout from "./components/layout/Layout";
import { AuthProvider } from "./context/AuthContext";
import { InterfaceProvider } from "./context/InterfaceContext";
import { ModalProvider } from "./context/ModalContext";
import { MouseProvider } from "./context/MouseContext";
import { PostsProvider } from "./context/PostsContext";
import { RoomsProvider } from "./context/RoomsContext";
import { SocketProvider } from "./context/SocketContext";
import { UsersProvider } from "./context/UsersContext";

function App() {
  return (
    <InterfaceProvider>
      <MouseProvider>
        <AuthProvider>
          <ModalProvider>
            <SocketProvider>
              <PostsProvider>
                <RoomsProvider>
                  <UsersProvider>
                    <Layout />
                  </UsersProvider>
                </RoomsProvider>
              </PostsProvider>
            </SocketProvider>
          </ModalProvider>
        </AuthProvider>
      </MouseProvider>
    </InterfaceProvider>
  );
}

export default App;
