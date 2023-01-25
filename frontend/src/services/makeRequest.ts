import axios, { RawAxiosRequestConfig } from "axios";

export const baseURL =
  process.env.NODE_ENV === "development"
    ? "http://localhost:8080"
    : "https://go-social-media-js.herokuapp.com";

const api = axios.create({
  baseURL,
});

export function makeRequest(url: string, options?: RawAxiosRequestConfig) {
  return api(url, options)
    .then((res: any) => res.data)
    .catch((e: any) =>
      Promise.reject((e.response?.data.msg ?? "Error").replace("Error: ", ""))
    );
}
