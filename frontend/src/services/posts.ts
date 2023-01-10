import type { CancelToken } from "axios";
import { baseURL, makeRequest } from "./makeRequest";

export type SortOrder = "DESC" | "ASC";
export type SortMode = "DATE" | "POPULARITY";

const getPage = (pageNum: number, order: SortOrder, mode: SortMode) =>
  makeRequest(
    `/api/posts/page/${pageNum}${
      order || mode
        ? `?${mode ? "mode=" + mode : ""}` +
          (order ? `${mode ? "&" : "?"}order=` + order : "")
        : ""
    }`,
    {
      method: "GET",
      withCredentials: true,
    }
  );

const submitComment = (comment: string, postId: string, parentId: string) =>
  makeRequest(
    `/api/posts/${postId}/comment${parentId && "?parent_id=" + parentId}`,
    {
      method: "POST",
      withCredentials: true,
      data: {
        content: comment,
      },
    }
  );

const deleteComment = (postId: string, commentId: string) =>
  makeRequest(`/api/posts/${postId}/comment/${commentId}/delete`, {
    method: "DELETE",
    withCredentials: true,
  });

const updateComment = (postId: string, commentId: string, comment: string) =>
  makeRequest(`/api/posts/${postId}/comment/${commentId}/update`, {
    method: "PATCH",
    withCredentials: true,
    data: { content: comment },
  });

const getPost = (slug: string) =>
  makeRequest(`/api/posts/${slug}`, { withCredentials: true }).then((data) => ({
    ...data,
    img_url: `${baseURL}/api/posts/${data.ID}/image?v=1`,
  }));

const getPostThumb = async (id: string, cancelToken: CancelToken) => {
  const data = await makeRequest(`/api/posts/${id}/thumb`, {
    responseType: "arraybuffer",
    cancelToken,
  });
  const blob = new Blob([data], { type: "image/jpeg" });
  return URL.createObjectURL(blob);
};

const getPostImage = async (id: string, cancelToken?: CancelToken) => {
  const data = await makeRequest(`/api/posts/${id}/image`, {
    responseType: "arraybuffer",
    cancelToken,
  });
  const blob = new Blob([data], { type: "image/jpeg" });
  return URL.createObjectURL(blob);
};

const getPostImageFile = async (id: string) => {
  const data = await makeRequest(`/api/posts/${id}/image`, {
    responseType: "arraybuffer",
  });
  return new File([data], "image.jpeg", { type: "image/jpeg" });
};

const createPost = (data: any) =>
  makeRequest(`/api/posts`, {
    withCredentials: true,
    method: "POST",
    data,
  });

const updatePost = (data: any, slug: string) =>
  makeRequest(`/api/posts/${slug}/update`, {
    withCredentials: true,
    method: "PATCH",
    data,
  });

const uploadPostImage = (file: File, slug: string) => {
  const data = new FormData();
  data.append("file", file);
  return makeRequest(`/api/posts/${slug}/image`, {
    method: "POST",
    withCredentials: true,
    data,
  });
};

const deletePost = (slug: string) =>
  makeRequest(`/api/posts/${slug}/delete`, {
    method: "DELETE",
    withCredentials: true,
  });

export {
  getPost,
  createPost,
  updatePost,
  uploadPostImage,
  getPage,
  getPostThumb,
  getPostImage,
  deletePost,
  getPostImageFile,
  submitComment,
  deleteComment,
  updateComment,
};
