import axios, { AxiosRequestConfig } from "axios";

export const baseURL =
  process.env.NODE_ENV === "development"
    ? "http://localhost:8080"
    : "https://go-social-media-js.herokuapp.com";

const api = axios.create({
  baseURL,
});

export function makeRequest(url: string, options?: AxiosRequestConfig) {
  return api(url, options)
    .then((res) => res.data)
    .catch((error) =>
      Promise.reject(
        (error.response?.data.msg ?? "Error").replace("Error: ", "")
      )
    );
}
