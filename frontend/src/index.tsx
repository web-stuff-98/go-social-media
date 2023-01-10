import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";

import { Routes, Route, BrowserRouter } from "react-router-dom";
import Home from "./routes/Home";
import Login from "./routes/Login";
import Register from "./routes/Register";
import NotFound from "./routes/NotFound";
import Policy from "./routes/Policy";
import Settings from "./routes/Settings";
import Editor from "./routes/Editor";
import Blog from "./routes/Blog";
import Page from "./routes/Page";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);
root.render(
  <BrowserRouter>
    <Routes>
      <Route path="/" element={<App />}>
        <Route index element={<Home />} />
        <Route path="post/:slug" element={<Page />} />
        <Route path="login" element={<Login />} />
        <Route path="blog/:page" element={<Blog />} />
        <Route path="policy" element={<Policy />} />
        <Route path="settings" element={<Settings />} />
        <Route path="register" element={<Register />} />
        <Route path="editor" element={<Editor />} />
        <Route path="editor/:slug" element={<Editor />} />
        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  </BrowserRouter>
);
