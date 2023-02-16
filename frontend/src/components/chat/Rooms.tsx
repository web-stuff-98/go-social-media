import { useEffect, useState, useCallback, useMemo, useRef } from "react";
import type { ChangeEvent } from "react";
import { BsChevronLeft, BsChevronRight } from "react-icons/bs";
import { FaSearch } from "react-icons/fa";
import useSocket from "../../context/SocketContext";
import { getRoomPage } from "../../services/rooms";
import classes from "../../styles/components/chat/Rooms.module.scss";
import { instanceOfChangeData } from "../../utils/DetermineSocketEvent";
import IconBtn from "../shared/IconBtn";
import ResMsg from "../shared/ResMsg";
import RoomCard from "./RoomCard";
import { debounce } from "lodash";
import { IRoomCard } from "../../interfaces/ChatInterfaces";
import { IResMsg } from "../../interfaces/GeneralInterfaces";
import type { CancelToken, CancelTokenSource } from "axios";
import axios from "axios";
import Dropdown from "../shared/Dropdown";

const roomDropdownItems = [
  { name: "OWN_ROOMS", node: "Your rooms" },
  { name: "INVITED_ROOMS", node: "Invited rooms" },
];

export default function Rooms() {
  const { socket, openSubscription, closeSubscription } = useSocket();

  const [roomDropdownIndex, setRoomDropdownIndex] = useState(0);
  const [pageNum, setPageNum] = useState(1);
  const [page, setPage] = useState<IRoomCard[]>([]);
  const [count, setCount] = useState(0);
  const [searchInput, setSearchInput] = useState("");
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const cancelSource = useRef<CancelTokenSource>();
  const cancelToken = useRef<CancelToken>();

  useEffect(() => {
    setResMsg({ msg: "", err: false, pen: true });
    const controller = new AbortController();
    handleSearch();
    return () => {
      controller.abort();
    };
    // eslint-disable-next-line
  }, [pageNum, searchInput, roomDropdownIndex]);

  const handleSearch = useMemo(
    () => debounce(() => updatePage(), 300),
    // eslint-disable-next-line
    [searchInput, page, roomDropdownIndex]
  );

  const updatePage = async () => {
    try {
      if (cancelToken.current) {
        cancelSource.current?.cancel();
      }
      cancelSource.current = axios.CancelToken.source();
      cancelToken.current = cancelSource.current?.token;
      const { count, rooms } = await getRoomPage(
        pageNum,
        searchInput,
        roomDropdownItems[roomDropdownIndex].name as
          | "OWN_ROOMS"
          | "INVITED_ROOMS",
        cancelToken.current!
      );
      setCount(Number(count));
      setPage(JSON.parse(rooms) as IRoomCard[]);
      setResMsg({ msg: "", err: false, pen: false });
    } catch (e) {
      if (!axios.isCancel(e)) {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      } else {
        setResMsg({ msg: "", err: false, pen: false });
      }
    }
  };

  useEffect(() => {
    openSubscription("room_feed");
    return () => {
      closeSubscription("room_feed");
    };
    // eslint-disable-next-line
  }, []);

  const nextPage = () => {
    setPageNum(Math.min(pageNum + 1, Math.ceil(count / 20)));
  };

  const prevPage = () => {
    setPageNum(Math.max(pageNum - 1, 1));
  };

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "ROOM") {
        updatePage();
      }
    }
    // eslint-disable-next-line
  }, []);

  useEffect(() => {
    if (socket) socket?.addEventListener("message", handleMessage);
    return () => {
      if (socket) socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  return (
    <div className={classes.container}>
      <div data-testid="Rooms list container" className={classes.rooms}>
        {page && !resMsg.pen && page.map((r) => <RoomCard key={r.ID} r={r} />)}
        <div
          data-testid="Rooms list ResMsg container"
          className={classes.resMsg}
        >
          <ResMsg resMsg={resMsg} />
        </div>
      </div>
      <form name="Search rooms" className={classes.searchContainer}>
        <input
          data-testid="Search room name input"
          name="Search room name"
          value={searchInput}
          onChange={(e: ChangeEvent<HTMLInputElement>) =>
            setSearchInput(e.target.value)
          }
          type="text"
          placeholder="Search rooms..."
        />
        <IconBtn
          testid="Search room button"
          type="submit"
          Icon={FaSearch}
          ariaLabel="Search rooms"
          name="Search rooms"
        />
      </form>
      <div className={classes.ownRoomsToggleAndPaginationControls}>
        {/*<Toggle
          toggledOn={onlyOwnRooms}
          setToggledOn={setOnlyOwnRooms}
          label="Your rooms"
        />
        <Toggle
          toggledOn={onlyOwnRooms}
          setToggledOn={setOnlyOwnRooms}
          label="Joined rooms"
        />*/}
        <Dropdown
          index={roomDropdownIndex}
          setIndex={setRoomDropdownIndex}
          light
          items={roomDropdownItems}
        />
        <div
          data-testid="Pagination controls container"
          className={classes.paginationControls}
        >
          <IconBtn
            onClick={prevPage}
            name="Prev page"
            ariaLabel="Previous page"
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
    </div>
  );
}
