import { useEffect, useState, useCallback, useMemo } from "react";
import type { ChangeEvent } from "react";
import { BsChevronLeft, BsChevronRight } from "react-icons/bs";
import { FaSearch } from "react-icons/fa";
import useSocket from "../../../context/SocketContext";
import { getRoomPage } from "../../../services/rooms";
import classes from "../../../styles/components/chat/Rooms.module.scss";
import { instanceOfChangeData } from "../../../utils/DetermineSocketEvent";
import IconBtn from "../../IconBtn";
import ResMsg, { IResMsg } from "../../ResMsg";
import RoomCard from "./RoomCard";
import { debounce } from "lodash";

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
  const [searchInput, setSearchInput] = useState("");
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  useEffect(() => {
    setResMsg({ msg: "", err: false, pen: true });
    handleSearch();
  }, [pageNum, searchInput]);

  const handleSearch = useMemo(
    () => debounce(() => updatePage(), 300),
    [searchInput, page]
  );

  const updatePage = () => {
    getRoomPage(pageNum, searchInput)
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
    setPageNum(Math.min(pageNum + 1, Math.ceil(count / 30)));
  };

  const prevPage = () => {
    setPageNum(Math.max(pageNum - 1, 1));
  };

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "ROOM") {
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
        {page && page.map((r) => <RoomCard key={r.ID} r={r} />)}
        <div className={classes.resMsg}>
          <ResMsg resMsg={resMsg} />
        </div>
      </div>
      <form className={classes.searchContainer}>
        <input
          value={searchInput}
          onChange={(e: ChangeEvent<HTMLInputElement>) =>
            setSearchInput(e.target.value)
          }
          type="text"
          placeholder="Search rooms..."
        />
        <IconBtn Icon={FaSearch} ariaLabel="Search rooms" name="Search rooms" />
      </form>
      <div className={classes.paginationControls}>
        <IconBtn
          onClick={prevPage}
          name="Prev page"
          ariaLabel="Prev page"
          Icon={BsChevronLeft}
        />
        {pageNum}/{Math.ceil(count / 30)}
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
