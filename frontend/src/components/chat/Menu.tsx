import classes from "../../styles/components/chat/Menu.module.scss";
import { ChatSection, useChat } from "./Chat";

export default function Menu() {
  const { setSection } = useChat();

  const MenuButton = ({ section }: { section: ChatSection }) => (
    <button name={section} onClick={() => setSection(section)}>
      {section}
    </button>
  );

  return (
    <div className={classes.container}>
      <MenuButton section={ChatSection.CONVS} />
      <MenuButton section={ChatSection.ROOMS} />
      <MenuButton section={ChatSection.EDITOR} />
    </div>
  );
}