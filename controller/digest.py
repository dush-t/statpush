import nnpy


def main():
    print("Listening on socket")
    sub = nnpy.Socket(nnpy.AF_SP, nnpy.SUB)
    sub.connect('ipc:///tmp/bmv2-0-notifications.ipc')
    sub.setsockopt(nnpy.SUB, nnpy.SUB_SUBSCRIBE, '')

    while(True):
        msg = sub.recv()
        print("message recieved")
        print(msg)


if __name__ == '__main__':
    main()