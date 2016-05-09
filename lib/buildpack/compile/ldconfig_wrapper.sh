if ! [ -x "$(command -v ldconfig)" ]; then
  LDCONFIG_COMMAND="/sbin/ldconfig"
else
  LDCONFIG_COMMAND="$(command -v ldconfig)"
fi

LDCONFIG_DIR=~/ldconfig
LDCONFIG_SCRIPT=$LDCONFIG_DIR/ldconfig

mkdir -p $LDCONFIG_DIR
echo "#!/bin/bash" > $LDCONFIG_SCRIPT
echo "$LDCONFIG_COMMAND -C $LDCONFIG_DIR/ld.so.cache \$@" >> $LDCONFIG_SCRIPT
chmod 755 $LDCONFIG_SCRIPT
export PATH=$LDCONFIG_DIR:$PATH

cp /etc/ld.so.conf $LDCONFIG_DIR
echo "$HOME/libunwind/lib/">> $LDCONFIG_DIR/ld.so.conf
echo "$HOME/gettext/lib/">> $LDCONFIG_DIR/ld.so.conf
ldconfig -f $LDCONFIG_DIR/ld.so.conf