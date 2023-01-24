import axios, { AxiosRequestConfig } from "axios";

export const baseURL =
  process.env.NODE_ENV === "development"
    ? "http://localhost:8080"
    : "https://go-social-media-js.herokuapp.com";

const api = axios.create({
  baseURL,
});

// I have to use type "any" instead of "AxiosRequestConfig" because
// after updating axios it causes stupid errors
export function makeRequest(url: string, options?: any) {
  return api(url, options)
    .then((res:any) => res.data)
    .catch((e:any) =>
      Promise.reject(
        (e.response?.data.msg ?? "Error").replace("Error: ", "")
      )
    );
}
