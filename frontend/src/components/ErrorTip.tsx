export default function ErrorTip({ message }: { message: string }) {
  return (
    <div
      style={{
        position: "absolute",
        left: "0.666rem",
        top: "1.33rem",
        padding: "var(--padding)",
        background: "red",
        color: "white",
      }}
    >
      {message}
    </div>
  );
}
