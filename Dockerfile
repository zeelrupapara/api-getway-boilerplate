FROM alpine:latest 

WORKDIR /app

COPY ./cmd/vfxserver .

COPY ./pkg/authz/model.conf ./pkg/authz/model.conf
COPY ./pkg/smtp/report_template.html ./pkg/smtp/report_template.html
COPY ./pkg/smtp/demo_account_template.html ./pkg/smtp/demo_account_template.html
COPY ./pkg/smtp/change_password_template.html ./pkg/smtp/change_password_template.html

EXPOSE 8888
CMD [ "./vfxserver" ]