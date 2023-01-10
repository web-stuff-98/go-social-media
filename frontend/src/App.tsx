import Layout from "./components/layout/Layout";
import { AuthProvider } from "./context/AuthContext";
import { InterfaceProvider } from "./context/InterfaceContext";
import { ModalProvider } from "./context/ModalContext";
import { MouseProvider } from "./context/MouseContext";
import { PostsProvider } from "./context/PostsContext";
import { RoomsProvider } from "./context/RoomsContext";
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
              <ModalProvider>
                <PostsProvider>
                  <RoomsProvider>
                    <UsersProvider>
                      <Layout />
                    </UsersProvider>
                  </RoomsProvider>
                </PostsProvider>
              </ModalProvider>
            </AuthProvider>
          </UserdropdownProvider>
        </SocketProvider>
      </MouseProvider>
    </InterfaceProvider>
  );
}

export default App;
