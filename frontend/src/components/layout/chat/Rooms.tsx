import { useEffect, useState, useCallback } from "react";
import { BsChevronLeft, BsChevronRight } from "react-icons/bs";
import { FaSearch } from "react-icons/fa";
import useSocket from "../../../context/SocketContext";
import { getRoomPage } from "../../../services/rooms";
import classes from "../../../styles/components/chat/Rooms.module.scss";
import { instanceOfChangeData } from "../../../utils/DetermineSocketEvent";
import IconBtn from "../../IconBtn";
import ResMsg, { IResMsg } from "../../ResMsg";
import RoomCard from "./RoomCard";

export interface IRoomCard {
  ID: string;
  name: string;
  author_id: string;
  img_blur?: string;
  img_url?: string;
}

export default function Rooms() {
  const { socket, openSubscription, closeSubscription } = useSocket();

  const [pageNum, setPageNum] = useState(1);
  const [page, setPage] = useState<IRoomCard[]>([]);
  const [count, setCount] = useState(0);
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  useEffect(() => {
    setResMsg({ msg: "", err: false, pen: true });
    updatePage();
  }, [pageNum]);

  const updatePage = () => {
    getRoomPage(pageNum)
      .then(({ count, rooms }: { count: string; rooms: string }) => {
        setCount(Number(count));
        setPage(JSON.parse(rooms) as IRoomCard[]);
        setResMsg({ msg: "", err: false, pen: false });
      })
      .catch((e) => {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      });
  };

  useEffect(() => {
    openSubscription("room_feed");
    return () => {
      closeSubscription("room_feed");
    };
  }, []);

  const nextPage = () => {
    setPageNum(Math.min(pageNum + 1, Math.ceil(count / 20)));
  };

  const prevPage = () => {
    setPageNum(Math.max(pageNum - 1, 1));
  };

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "ROOM") {
        console.log("CHANGE");
        updatePage();
      }
    }
  }, []);

  useEffect(() => {
    if (socket) socket?.addEventListener("message", handleMessage);
    return () => {
      if (socket) socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  return (
    <div className={classes.container}>
      <div className={classes.rooms}>
        {page.map((r) => (
          <RoomCard r={r} />
        ))}
        <div className={classes.resMsg}>
          <ResMsg resMsg={resMsg} />
        </div>
      </div>
      <form className={classes.searchContainer}>
        <input type="text" placeholder="Search rooms..." />
        <IconBtn Icon={FaSearch} ariaLabel="Search rooms" name="Search rooms" />
      </form>
      <div className={classes.paginationControls}>
        <IconBtn
          onClick={prevPage}
          name="Prev page"
          ariaLabel="Prev page"
          Icon={BsChevronLeft}
        />
        {pageNum}/{Math.ceil(count / 20)}
        <IconBtn
          onClick={nextPage}
          name="Next page"
          ariaLabel="Next page"
          Icon={BsChevronRight}
        />
      </div>
    </div>
  );
}
