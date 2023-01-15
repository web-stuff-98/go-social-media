export default function ErrorTip({ message }: { message: string }) {
  return (
    <div
      style={{
        position: "absolute",
        left: "var(--padding-base)",
        bottom: "-1rem",
        background: "red",
        textAlign:"left",
        padding:"0 var(--padding-medium)",
        color: "white",
        zIndex:"99",
        filter:"opacity(0.8)",
        borderRadius:"var(--border-radius-medium)"
      }}
    >
      {message}
    </div>
  );
}
