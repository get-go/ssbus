FROM scratch
ADD main /

EXPOSE 3000

CMD ["/main"]
